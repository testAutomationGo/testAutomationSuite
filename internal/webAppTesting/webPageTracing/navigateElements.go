package webPageTracing

import (
	"testAutomationSuiteGO/internal/logger"

	"github.com/playwright-community/playwright-go"
)

func ClickOnElementByTagAndText(page playwright.Page, tag string, text string) error {
	elements, err := GetAllElementsStructured(page)
	if err != nil {
		logger.Log("Error getting elements: "+err.Error(), "")
		return err
	}
	for _, element := range elements {
		if element.Tag == tag && element.Text == text {
			selector := element.Selector
			if selector != "" {
				err := page.Locator(selector).Click()
				if err != nil {
					logger.Log("Error clicking on element: "+err.Error(), "")
				} else {
					return nil
				}
			} else {
				logger.Log("Selector is empty for element with tag: "+tag+" and text: "+text, "")
			}
			return nil
		}
	}
	return nil
}
