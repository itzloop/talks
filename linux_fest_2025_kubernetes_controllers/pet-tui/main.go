package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	tea "github.com/charmbracelet/bubbletea"
	v2025 "github.com/itzloop/pet-controller/api/v2025"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	scheme := runtime.NewScheme()
	if err := v2025.AddToScheme(scheme); err != nil {
		log.Fatalln("failed to add scheme", err)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Fatalln("Failed to create manager:", err)
	}


    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
    defer cancel()

	grp, ctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		// Start the informer cache in a background goroutine
		err := mgr.Start(ctx)
		if err != nil {
			log.Println("Manager error:", err)
		}

        cancel()
		return err
	})


    if !mgr.GetCache().WaitForCacheSync(ctx) {
       log.Fatalln("failed to sync the cache") 
    }

	tui := New(mgr.GetClient())
	p := tea.NewProgram(tui, tea.WithContext(ctx))

	grp.Go(func() error {
		// Create the TUI model with the client
		_, err := p.Run()
		if err != nil {
			log.Println("Error running TUI:", err)
		}

        cancel()
		return err
	})

	if err := grp.Wait(); err != nil {
		log.Println("one of the goroutines failed", err)
	}

}

// func main() {
// 	cfg, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
// 	if err != nil {
// 		log.Fatalf("could not load kubeconfig: %v\n", err)
// 	}

// 	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
// 		Scheme: scheme,
// 	})
// 	if err != nil {
// 		log.Fatalf("failed to create manager: %v\n", err)
// 	}

// 	crl := &notAPetController{
// 		Client: mgr.GetClient(),
// 	}

// 	p := tea.NewProgram(initialModel(crl))
// 	if err := setupNotAPetController(context.Background(), p, mgr, crl); err != nil {
// 		panic(err)
// 	}

// 	fmt.Println("👀 Watching pets...")

// 	go func() {
// 		if err := mgr.Start(context.Background()); err != nil {
// 			log.Fatalf("Error running manager: %v\n", err)
// 		}
// 	}()

// 	managerStarted := &sync.WaitGroup{}
// 	managerStarted.Add(1)
// 	mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
// 		// Your function here
// 		managerStarted.Done()
// 		fmt.Println("🎉 Manager started, now doing something fun!")
// 		return nil
// 	}))

// 	fmt.Println("hi")
// 	managerStarted.Wait()

// 	p.Send(petUpdateMsg{})
// 	fmt.Println("hello")
// 	if _, err := p.Run(); err != nil {
// 		log.Fatalf("Error running app: %v\n", err)
// 	}

// }
