package playwrightInternal

import (
	"log"

	"github.com/playwright-community/playwright-go"
)

func InitPlaywrightWithContextUtil(headless bool, profile string, contextOptions playwright.BrowserNewContextOptions) (*playwright.Playwright, playwright.Browser, playwright.BrowserContext, playwright.Page) {
	pw, err := playwright.Run()
	if err != nil {
		log.Println("Error starting playwright: " + err.Error())
		return nil, nil, nil, nil
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		//Channel: playwright.String("chrome")
		Headless: playwright.Bool(headless),
	})
	if err != nil {
		log.Println("Error launching browser: " + err.Error())
		return nil, nil, nil, nil
	}
	context, err := browser.NewContext(contextOptions)
	if err != nil {
		log.Println("Error creating context: " + err.Error())
		return nil, nil, nil, nil
	}
	page, err := context.NewPage()
	if err != nil {
		log.Println("Error creating page: " + err.Error())
		return nil, nil, nil, nil
	}
	SetTimeoutByProfile(page, profile)
	return pw, browser, context, page
}

func GoToUtil(page playwright.Page, url string) {
	_, err := page.Goto(url)
	if err != nil {
		log.Println("Could not navigate to URL: " + err.Error())
		return
	}
}
