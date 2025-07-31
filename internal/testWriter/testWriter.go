package testWriter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testAutomationSuiteGO/internal/testingToolkit"
	"testAutomationSuiteGO/internal/webAppTesting"
	"time"

	"github.com/chromedp/chromedp"
)

type Action struct {
	Type     string `json:"type"`
	Selector string `json:"selector"`
	Value    string `json:"value,omitempty"`
}

var (
	RecordedActions []Action
	mu              sync.Mutex
)

func addAction(action Action) {
	mu.Lock()
	defer mu.Unlock()
	RecordedActions = append(RecordedActions, action)
}

func getRecordedActions() []Action {
	mu.Lock()
	defer mu.Unlock()
	return RecordedActions
}

func clearActions() {
	mu.Lock()
	defer mu.Unlock()
	RecordedActions = nil
}

func saveActions(actions []Action) error {
	fmt.Printf("Saving %d actions\n", len(actions))
	for _, action := range actions {
		fmt.Printf("%+v\n", action)
	}
	return nil
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func RecordActionsInTheBrowser(url string) []Action {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
	)
	allocatorCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocatorCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		fmt.Println("Received event request")
		var action Action
		if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
			fmt.Println("Error decoding action:", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		addAction(action)
	})

	server := &http.Server{Addr: ":8080"}
	go func() {
		fmt.Println("Starting HTTP server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Evaluate(`(function() {
			function sendEvent(action) {
				console.log("Sending event: ", action);
				fetch('http://localhost:8080/event', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(action)
				})
				.then(response => response.text())
				.then(data => console.log("Event sent: ", data))
				.catch(error => console.error("Error sending event: ", error));
			}

			let inputTimeout;
			document.addEventListener('input', function(event) {
				clearTimeout(inputTimeout);
				inputTimeout = setTimeout(function() {
					var element = event.target;
					var selector = element.tagName.toLowerCase();
					if (element.id) {
						selector += "#" + element.id;
					}
					if (element.className) {
						selector += "." + element.className.replace(/\s+/g, '.');
					}
					var action = { type: 'input', selector: selector, value: element.value };
					sendEvent(action);
					console.log("Input event: ", action);
				}, 500); // Adjust timeout as needed
			});

			document.addEventListener('click', function(event) {
				var element = event.target;
				var selector = element.tagName.toLowerCase();
				if (element.id) {
					selector += "#" + element.id;
				}
				if (element.className) {
					selector += "." + element.className.replace(/\s+/g, '.');
				}
				var action = { type: 'click', selector: selector };
				sendEvent(action);
				console.log("Click event: ", action);
			});
		})()`, nil),
	)
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	select {
	case <-c:
		fmt.Println("Received interrupt, stopping...")
	case <-ctx.Done():
		fmt.Println("Browser context done, stopping...")
	}
	time.Sleep(time.Second * 2)

	if err := server.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP server Shutdown: %v", err)
	}

	actions := getRecordedActions()

	if err := saveActions(actions); err != nil {
		log.Printf("Error saving actions: %v", err)
	}

	clearActions()

	return actions
}

func ReplayActions(actions []Action, beginningURL string) error {

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
	)
	allocatorCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocatorCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	if err := chromedp.Run(ctx, chromedp.Navigate(beginningURL)); err != nil {
		return err
	}

	for _, action := range actions {
		switch action.Type {
		case "pageload":
			webAppTesting.Wait(ctx, "", "")
		case "click":
			if err := chromedp.Run(ctx, chromedp.Click(action.Selector)); err != nil {
				return err
			}
		case "input":
			if err := chromedp.Run(ctx, chromedp.SendKeys(action.Selector, action.Value)); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown action type: %s", action.Type)
		}
		testingToolkit.DelayMilliseconds(500)
	}
	return nil
}
