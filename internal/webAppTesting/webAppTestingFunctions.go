package webAppTesting

import (
	"context"
	internalChromedp "testAutomationSuiteGO/internal/chromedpInternal"
	"testAutomationSuiteGO/internal/logger"
)

func NavigateTo(ctx context.Context, url string, tcNumber string) {
	navigated := internalChromedp.NavigateToURL(ctx, url)
	if !navigated {
		logger.Log("Navigation Error", tcNumber)
	}
}

func Click(ctx context.Context, selector string, tcNumber string) {
	clicked := internalChromedp.ClickOnSelector(ctx, selector)
	if !clicked {
		logger.Log("Click Error On: "+selector, tcNumber)
	}
}

func Wait(ctx context.Context, selector string, tcNumber string) {
	waited := internalChromedp.WaitOnSelector(ctx, selector)
	if !waited {
		logger.Log("Wait Error On: "+selector, tcNumber)
	}
}

func SendKeys(ctx context.Context, selector string, keys string, tcNumber string) {
	sent := internalChromedp.SendKeysToSelector(ctx, selector, keys)
	if !sent {
		logger.Log("SendKeys Error On: "+selector, tcNumber)
	}
}

func GetText(ctx context.Context, selector string, tcNumber string) string {
	text, gotText := internalChromedp.GetTextFromSelector(ctx, selector)
	if !gotText {
		logger.Log("GetText Error On: "+selector, tcNumber)
	}
	return text
}

func WaitUntilInvisible(ctx context.Context, selector string, tcNumber string) {
	waited := internalChromedp.WaitUntilInvisible(ctx, selector)
	if !waited {
		logger.Log("WaitUntilInvisible Error On: "+selector, tcNumber)
	}
}

func GetTitle(ctx context.Context, tcNumber string) string {
	title, gotTitle := internalChromedp.GetTitle(ctx)
	if !gotTitle {
		logger.Log("GetTitle Error", tcNumber)
	}
	return title
}

func RefreshPage(ctx context.Context, tcNumber string) {
	refreshed := internalChromedp.RefreshPage(ctx)
	if !refreshed {
		logger.Log("RefreshPage Error", tcNumber)
	}
}

func GetSource(ctx context.Context, tcNumber string) string {
	source, _ := internalChromedp.GetSource(ctx)
	return source
}

func VerifyElementExists(ctx context.Context, selector, tcNumber string) bool {
	exists := internalChromedp.ElementExists(ctx, selector)
	if !exists {
		logger.Log("Element Does Not Exist: "+selector, tcNumber)
	}
	return true
}

func MaximizeBrowser(ctx context.Context, tcNumber string) {
	maximized := internalChromedp.MaximizeBrowser(ctx)
	if !maximized {
		logger.Log("MaximizeBrowser Error", tcNumber)
	}
}

func WaitUntilEnabled(ctx context.Context, selector, tcNumber string) {
	waited := internalChromedp.WaitUntilSelectorIsEnabled(ctx, selector)
	if !waited {
		logger.Log("WaitUntilEnabled Error On: "+selector, tcNumber)
	}
}

func ClickOnText(ctx context.Context, text, tcNumber string) {
	clicked := internalChromedp.ClickOnElementByText(ctx, text)
	if !clicked {
		logger.Log("ClickOnText Error On: "+text, tcNumber)
	}
}

func SelectOption(ctx context.Context, selector, option, tcNumber string) {
	selected := internalChromedp.SelectFromADropDown(ctx, selector, option)
	if !selected {
		logger.Log("SelectOption Error On: "+selector, tcNumber)
	}
}
