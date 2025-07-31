package chromedpInternal

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testAutomationSuiteGO/internal/testRunParameters"
	"testAutomationSuiteGO/internal/testingToolkit"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
)

func NonHeadlessBrowserContextForNonAutomatedTests() (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false), // Run browser in non-headless mode
	)
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	// No timeout is set for ctx, operations will run indefinitely until complete or manually cancelled
	return ctx, cancel
}

func NonHeadlessBrowserContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
	)
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	return ctxWithTimeout, cancel
}

// Must Set EnvInt before calling this function
func GetBrowserContext(timeOut int) (context.Context, context.CancelFunc) {
	timeDuration := time.Duration(timeOut) * time.Second
	if testRunParameters.GetRunLocal() {
		ctx, cancel := NonHeadlessBrowserContext(timeDuration)
		return ctx, cancel
	}
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-software-rasterizer", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("single-process", true),
	)
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeDuration)
	finalContext, _ := chromedp.NewContext(ctxWithTimeout)
	return finalContext, cancel
}

func CloseBrowser(ctx context.Context) bool {
	err := chromedp.Run(ctx, ClosingBrowser())
	if err != nil {
		println(err.Error())
		return false
	}
	return true
}

func ClosingBrowser() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		err := browser.Close().Do(ctx)
		return err
	}
}

func NavigateToURL(ctx context.Context, url string) bool {
	err := chromedp.Run(ctx, chromedp.Navigate(url))
	if err != nil {
		println(err.Error())
		return false
	}
	return true
}

func RefreshPage(ctx context.Context) bool {
	err := chromedp.Run(ctx, chromedp.Reload())
	if err != nil {
		println(err.Error())
		return false
	}
	return true
}

func WaitOnSelector(ctx context.Context, selector string) bool {
	err := chromedp.Run(ctx, chromedp.WaitVisible(selector))
	if err != nil {
		println("Wait On Selector Error: " + err.Error())
		return false
	}
	return true
}

func ClickOnSelector(ctx context.Context, selector string) bool {
	err := chromedp.Run(ctx, chromedp.Click(selector))
	if err != nil {
		println(err.Error())
		return false
	}
	return true
}

func SendKeysToSelector(ctx context.Context, selector string, keys string) bool {
	err1 := chromedp.Run(ctx, chromedp.SendKeys(selector, keys))
	if err1 != nil {
		println(err1.Error())
		return false
	}
	return true
}

func GetTextFromSelector(ctx context.Context, selector string) (string, bool) {
	var text string
	err := chromedp.Run(ctx, chromedp.Text(selector, &text))
	if err != nil {
		text = "Unable to get text from selector."
		println(err.Error())
		return text, false
	}
	return text, true
}

func WaitUntilInvisible(ctx context.Context, selector string) bool {
	err := chromedp.Run(ctx, chromedp.WaitNotVisible(selector))
	if err != nil {
		println("Wait Until Invisible Error: " + err.Error())
		return false
	}
	return true
}

func GetTitle(ctx context.Context) (string, bool) {
	var title string
	err := chromedp.Run(ctx, chromedp.Title(&title))
	if err != nil {
		println(err.Error())
		return "Unable to get title.", false
	}
	return title, true
}

func GetSource(ctx context.Context) (string, bool) {
	var source string
	err := chromedp.Run(ctx, chromedp.OuterHTML("html", &source))
	if err != nil {
		println("Get Source Error: " + err.Error())
		return "Unable to get source.", false
	}
	return source, true
}

func WaitUntilSelectorIsEnabled(ctx context.Context, selector string) bool {
	err := chromedp.Run(ctx, chromedp.WaitEnabled(selector))
	if err != nil {
		println(err.Error())
		return false
	}
	return true
}

func MaximizeBrowser(ctx context.Context) bool {
	err := chromedp.Run(ctx, chromedp.EmulateViewport(1920, 1080))
	if err != nil {
		println(err.Error())
		return false
	}
	return true
}

func VerifyStringInSource3SecondDelay(ctx context.Context, stringToVerify string) bool {
	source, _ := GetSource(ctx)
	contains := strings.Contains(source, stringToVerify)
	if contains {
		return true
	}
	for i := 0; i < 3; i++ {
		testingToolkit.DelaySeconds(1)
		source, _ = GetSource(ctx)
		contains = strings.Contains(source, stringToVerify)
		if contains {
			return true
		}
	}
	return strings.Contains(source, stringToVerify)
}

func WaitForBody(ctx context.Context) bool {
	err := chromedp.Run(ctx, chromedp.WaitVisible("body"))
	if err != nil {
		println("Wait for Body Error: " + err.Error())
		return false
	}
	return true
}

func ElementExists(ctx context.Context, selector string) bool {
	var exists bool
	err := chromedp.Run(ctx,
		chromedp.EvaluateAsDevTools(
			`document.querySelector(`+selector+`) !== null && document.querySelector(`+selector+`).offsetParent !== null`,
			&exists,
		),
	)
	if err != nil {
		log.Println("Element Exists Error:", err)
		return false
	}
	return exists
}

func ClickOnElementByText(ctx context.Context, text string) bool {
	excapedText := escapeQuotes(text)
	js := fmt.Sprintf(`document.evaluate("//text()[contains(., '%s')]/..", document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue.click()`, excapedText)
	err := chromedp.Run(ctx,
		chromedp.Evaluate(js, nil),
	)
	if err != nil {
		println(err.Error())
		return false
	}
	return true
}

func escapeQuotes(input string) string {
	return fmt.Sprintf("%s%s%s", "'", input, "'")
}

func SelectFromADropDown(ctx context.Context, selector string, value string) bool {
	err := chromedp.Run(ctx,
		chromedp.Click(selector),
		chromedp.Sleep(1*time.Second),
		chromedp.SetValue(selector, value),
	)
	if err != nil {
		println(err.Error())
		return false
	}
	return true
}
