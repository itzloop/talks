package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	v2025 "github.com/itzloop/pet-controller/api/v2025"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Pet v2025.Pet

func (p Pet) Emoji() string {
	switch {
	case p.Status.Food == 0:
		return "💀"
	case p.Status.Food < 30 && p.Status.Love == 0:
		return "🤬"
	case p.Status.Food >= 30 && p.Status.Love == 0:
		return "😭"
	case p.Status.Food < 30 || p.Status.Love < 30:
		return "😠"
	case p.Status.Food < 50 || p.Status.Love < 50:
		return "😢"
	case p.Status.Love > 90 && p.Status.Food > 80:
		return "🥰"
	case p.Status.Food >= 80 && p.Status.Love >= 80:
		return "😍"
	default:
		return "🙂"
	}
}

// Define the model struct in one place
type model struct {
	k8s    client.Client
	pets   []v2025.Pet
	cursor int
	err    error
}

func New(k8s client.Client) tea.Model {
	return &model{
		k8s: k8s,
	}
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Errorf("Error fetching pets: %w", m.err).Error()
	}
	if m.pets == nil {
		return "Loading pets..."
	}

	var b strings.Builder
	for i, p := range m.pets {
		cursor := "  "
		if i == m.cursor {
			cursor = "👉"
		}

		pet := Pet(p)
		fmt.Fprintf(&b, "%s %s  %s\n", cursor, pet.Emoji(), lipgloss.NewStyle().Bold(true).Render(p.Spec.Nickname))
		fmt.Fprintf(&b, "   🍗 Food: %s  (%d)\n", bar(pet.Status.Food), pet.Status.Food)
		fmt.Fprintf(&b, "   ❤️ Love: %s  (%d)\n\n", bar(pet.Status.Love), pet.Status.Love)
	}
	b.WriteString("⬆⬇: Move 🧭  |  f: Feed  🍗  |  l: Love  ❤️  |  q: Quit ❌\n")
	b.WriteString("             |  F: 🍗🍗🍗🍗  |  L: ❤️❤️❤️❤️  |            \n")
	return b.String()
}

type tickMsg struct{}
type errMsg struct{ error }

func (m model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return tickMsg{}
		},
	)
}

func SortPetsByAge(pets []v2025.Pet) {
	sort.Slice(pets, func(i, j int) bool {
		return pets[i].ObjectMeta.CreationTimestamp.Time.Before(
			pets[j].ObjectMeta.CreationTimestamp.Time,
		)
	})
}

func (m model) UpdatePet(petName string, ns string, food, love int) error {
	if food > 100 || love > 100 || food < 0 || love < 0 {
		return fmt.Errorf("invalid food=%d and love=%d values should be between 0 and 100", food, love)
	}

	ctx := context.Background()

	var pet v2025.Pet
	err := m.k8s.Get(ctx, client.ObjectKey{
		Name:      petName,
		Namespace: ns,
	}, &pet)
	if err != nil {
		return err
	}

	for range 10 {
		petCopy := pet.DeepCopy()
		// Modify the spec values
		if food != 0 {
			petCopy.Status.Food = food
			petCopy.Status.FedTime = v1.Now()
		}

		if love != 0 {
			petCopy.Status.Love = love
			petCopy.Status.PetTime = v1.Now()
		}

		if food == 0 && love == 0 {
			return nil
		}

        petCopy.Annotations["fed"] = "true"
        petCopy.Annotations["loved"] = "true"

		if err := m.k8s.Update(ctx, petCopy); err != nil {
			if errors.IsConflict(err) {
				if err := m.k8s.Get(ctx, client.ObjectKey{Name: petCopy.Name, Namespace: pet.Namespace}, &pet); err != nil {
					return err
				}

				continue
			} else if errors.IsNotFound(err) {
				return nil
			}

			return err
		}

		return nil
	}

	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		ctx := context.Background()
		var list v2025.PetList
		if err := m.k8s.List(ctx, &list); err != nil {
			return m, func() tea.Msg { return errMsg{err} }
		}

		SortPetsByAge(list.Items)
		m.pets = list.Items

		// Ensure cursor stays within bounds
		if len(m.pets) == 0 {
			m.cursor = 0
		} else if m.cursor >= len(m.pets) {
			m.cursor = len(m.pets) - 1
		}

		// do this one second later
		return m, tea.Tick(1*time.Second, func(time.Time) tea.Msg { return tickMsg{} })
	case errMsg:
		m.err = msg
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.pets)-1 {
				m.cursor++
			}

		case "f", "F":
			m.pets[m.cursor].Status.Food += 10
			if msg.String() == "F" {
				m.pets[m.cursor].Status.Food = 100
			}
			if m.pets[m.cursor].Status.Food > 100 {
				m.pets[m.cursor].Status.Food = 100
			}
			pet := m.pets[m.cursor]
			if err := m.UpdatePet(pet.Name, pet.Namespace, pet.Status.Food, 0); err != nil {
				return m, func() tea.Msg { return errMsg{fmt.Errorf("failed to update pet: %w", err)} }
			}
		case "l", "L":
			m.pets[m.cursor].Status.Love += 10
			if msg.String() == "L" {
				m.pets[m.cursor].Status.Love = 100
			}
			if m.pets[m.cursor].Status.Love > 100 {
				m.pets[m.cursor].Status.Love = 100
			}
			pet := m.pets[m.cursor]
			if err := m.UpdatePet(pet.Name, pet.Namespace, 0, pet.Status.Love); err != nil {
				return m, func() tea.Msg { return errMsg{fmt.Errorf("failed to update pet: %w", err)} }
			}
		}
	}
	return m, nil
}

// bar function is moved to view.go
// model struct is defined in update.go
