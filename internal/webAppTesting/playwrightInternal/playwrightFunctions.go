package playwrightInternal

import (
	"fmt"
	"os"
	"strings"
	"testAutomationSuiteGO/internal/logger"
	"testAutomationSuiteGO/internal/testingToolkit"

	"github.com/playwright-community/playwright-go"
)

func InitPlaywright() (*playwright.Playwright, playwright.Browser, playwright.Page, error) {
	pw, err := playwright.Run()
	if err != nil {
		logger.Log("could not start playwright: "+err.Error(), "")
		return nil, nil, nil, err
	}
	browser, err := pw.Chromium.Launch()
	if err != nil {
		logger.Log("could not launch browser: "+err.Error(), "")
		return nil, nil, nil, err
	}
	page, err := browser.NewPage()
	if err != nil {
		logger.Log("could not create page: "+err.Error(), "")
		return nil, nil, nil, err
	}
	page.SetDefaultTimeout(10000)
	page.SetDefaultNavigationTimeout(10000)
	return pw, browser, page, nil
}

func Run(tcNumber string) (*playwright.Playwright, error) {
	pw, err := playwright.Run()
	if err != nil {
		logger.Log("could not start playwright: "+err.Error(), tcNumber)
		return nil, err
	}
	return pw, nil
}

func Browser(pw *playwright.Playwright, tcNumber string) (playwright.Browser, error) {
	browser, err := pw.Chromium.Launch()
	if err != nil {
		logger.Log("could not launch browser: "+err.Error(), tcNumber)
		return nil, err
	}
	return browser, nil
}

func Page(browser playwright.Browser, tcNumber string) (playwright.Page, error) {
	page, err := browser.NewPage()
	if err != nil {
		logger.Log("could not create page: "+err.Error(), tcNumber)
		return nil, err
	}
	page.SetDefaultTimeout(20000)
	page.SetDefaultNavigationTimeout(20000)
	return page, nil
}

func GoTo(page playwright.Page, url string, tcNumber string) error {
	_, err := page.Goto(url)
	if err != nil {
		logger.Log("could not navigate to url: "+err.Error(), tcNumber)
		return err
	}
	return nil
}

func Refresh(page playwright.Page, tcNumber string) error {
	_, err := page.Reload()
	if err != nil {
		logger.Log("could not refresh page: "+err.Error(), tcNumber)
		return err
	}
	return nil
}

func Close(page playwright.Page, tcNumber string) error {
	err := page.Close()
	if err != nil {
		logger.Log("could not close page: "+err.Error(), tcNumber)
		return err
	}
	return nil
}

func Click(page playwright.Page, selector string, tcNumber string) error {
	selectorLocator := page.Locator(selector)
	if err := selectorLocator.Click(); err != nil {
		logger.Log("could not click on selector: "+err.Error(), tcNumber)
		return err
	}
	return nil
}

func FillTextBox(page playwright.Page, selector string, text string, tcNumber string) error {
	selectorLocator := page.Locator(selector)
	if err := selectorLocator.Fill(text); err != nil {
		logger.Log("could not fill text box: "+err.Error(), tcNumber)
		return err
	}
	return nil
}

func GetText(page playwright.Page, selector string, tcNumber string) (string, error) {
	selectorLocator := page.Locator(selector)
	text, err := selectorLocator.TextContent()
	if err != nil {
		logger.Log("could not get text: "+err.Error(), tcNumber)
		return "", err
	}
	return text, nil
}

func GetTitle(page playwright.Page, tcNumber string) (string, error) {
	title, err := page.Title()
	if err != nil {
		logger.Log("could not get title: "+err.Error(), tcNumber)
		return "", err
	}
	return title, nil
}

func GetSource(page playwright.Page, tcNumber string) (string, error) {
	source, err := page.Content()
	if err != nil {
		logger.Log("could not get source: "+err.Error(), tcNumber)
	}
	return source, nil
}

func SelectOption(page playwright.Page, selector string, option string, tcNumber string) error {
	selectorLocator := page.Locator(selector)
	optionsSlice := []string{option}
	if _, err := selectorLocator.SelectOption(playwright.SelectOptionValues{Values: &optionsSlice}, playwright.LocatorSelectOptionOptions{Timeout: playwright.Float(5000)}); err != nil {
		logger.Log("could not select option: "+err.Error(), tcNumber)
		return err
	}
	return nil
}

func GetByText(page playwright.Page, text string, tcNumber string) (playwright.Locator, error) {
	selectorLocator := page.Locator("button:has(p:text(\"" + text + "\"))")
	if err := selectorLocator.WaitFor(); err != nil {
		logger.Log("could not get by text: "+err.Error(), tcNumber)
		return nil, err
	}
	return selectorLocator, nil
}

func VerifyStringPresentInSourceXSecondDelay(page playwright.Page, text string, seconds int, tcNumber string) bool {
	testingToolkit.DelaySeconds(seconds)
	source, err := page.Content()
	if err != nil {
		logger.Log("could not get source: "+err.Error(), tcNumber)
	}
	var isPresent bool
	if source != "" {
		if strings.Contains(source, text) {
			isPresent = true
		} else {
			isPresent = false
		}
	} else {
		logger.Log("source is empty, cannot verify string presence", tcNumber)
		isPresent = false
	}
	return isPresent
}

func ClickOnElementByTypeAndId(page playwright.Page, elementType, id, tcNumber string) error {
	selector := fmt.Sprintf("%s#%s", elementType, id)

	err := page.Locator(selector).Click()
	if err != nil {
		return fmt.Errorf("failed to click on %s with id '%s': %w", elementType, id, err)
	}
	return nil
}

func ClickButtonBySelectorAndButtonIndexWithJS(page playwright.Page, selector string, buttonIndex int, tcNumber string) error {
	jsCode := fmt.Sprintf(`
        () => {
            const buttons = document.querySelectorAll('%s');
            if (buttons.length > %d) {
                const button = buttons[%d];
                
                // Log element details
                console.log('Clicking button:', button);
                console.log('Button offsetParent:', button.offsetParent);
                console.log('Button style:', window.getComputedStyle(button));
                
                // Try different click methods
                try {
                    button.click();
                    return 'clicked with click()';
                } catch (e1) {
                    try {
                        button.dispatchEvent(new MouseEvent('click', {
                            view: window,
                            bubbles: true,
                            cancelable: true
                        }));
                        return 'clicked with dispatchEvent';
                    } catch (e2) {
                        return 'both click methods failed: ' + e1.message + ', ' + e2.message;
                    }
                }
            }
            return 'button not found';
        }
    `, selector, buttonIndex, buttonIndex)

	result, err := page.Evaluate(jsCode)
	if err != nil {
		return fmt.Errorf("JavaScript execution failed: %w", err)
	}

	logger.Log(fmt.Sprintf("JavaScript click result: %v", result), tcNumber)
	return nil
}

func ClickElementBySelectorWithJS(page playwright.Page, selector, tcNumber string) error {
	jsCode := fmt.Sprintf(`
        () => {
            const buttons = document.querySelectorAll('%s');
            if (buttons.length > 0) {
                for (let i = 0; i < buttons.length; i++) {
                    const button = buttons[i];
                    const rect = button.getBoundingClientRect();
                    const style = window.getComputedStyle(button);
                    
                    if (rect.width > 0 && rect.height > 0 && 
                        style.display !== 'none' && 
                        style.visibility !== 'hidden') {
                        
                        console.log('Clicking button at index:', i);
                        button.click();
                        return 'success: clicked button ' + i;
                    }
                }
                return 'no clickable buttons found';
            }
            return 'no buttons found';
        }
    `, selector)

	result, err := page.Evaluate(jsCode)
	if err != nil {
		return fmt.Errorf("JavaScript execution failed: %w", err)
	}

	logger.Log(fmt.Sprintf("JavaScript click result: %v", result), tcNumber)
	return nil
}

func ClickByTypeAttributeValue(page playwright.Page, elementType, attribute, attributeValue, tcNumber string) error {
	selector := fmt.Sprintf("%s[%s='%s']", elementType, attribute, attributeValue)

	selectorLocator := page.Locator(selector)
	if err := selectorLocator.Click(); err != nil {
		logger.Log("could not click on selector: "+err.Error(), tcNumber)
		return fmt.Errorf("failed to click on %s with %s='%s': %w", elementType, attribute, attributeValue, err)
	}
	return nil
}

func ClickByText(page playwright.Page, text, tcNumber string) error {
	selector := fmt.Sprintf("text=%s", text)
	selectorLocator := page.Locator(selector)
	if err := selectorLocator.Click(); err != nil {
		logger.Log("could not click on element by text: "+err.Error(), tcNumber)
		return fmt.Errorf("failed to click on element with text '%s': %w", text, err)
	}
	return nil
}

func ClickFirstByText(page playwright.Page, text, tcNumber string) error {
	selector := fmt.Sprintf("text=%s", text)
	locators := page.Locator(selector)
	count, err := locators.Count()
	if err != nil {
		logger.Log("could not count elements by text: "+err.Error(), tcNumber)
		return fmt.Errorf("failed to count elements with text '%s': %w", text, err)
	}
	if count == 0 {
		return fmt.Errorf("no elements found with text '%s'", text)
	}
	if err := locators.Nth(0).Click(); err != nil {
		logger.Log("could not click on first element by text: "+err.Error(), tcNumber)
		return fmt.Errorf("failed to click on first element with text '%s': %w", text, err)
	}
	return nil
}

func DownloadAndVerify(page playwright.Page, buttonText, savePath string, tcNumber string) (bool, error) {
	download, err := page.ExpectDownload(func() error {
		return page.Locator("text=" + buttonText).Click()
	})
	if err != nil {
		return false, err
	}
	if err := download.SaveAs(savePath); err != nil {
		return false, err
	}
	if _, err := os.Stat(savePath); err == nil {
		return true, nil
	}
	logger.Log("downloaded file does not exist: "+savePath, tcNumber)
	return false, err
}
