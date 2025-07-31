package webPageTracing

import (
	"encoding/json"
	"fmt"
	"testAutomationSuiteGO/internal/logger"
	"testAutomationSuiteGO/internal/testRunParameters"
	"testAutomationSuiteGO/internal/testingToolkit"

	"github.com/playwright-community/playwright-go"
)

type Element struct {
	Tag        string            `json:"tag"`
	ID         string            `json:"id"`
	Text       string            `json:"text"`
	XPath      string            `json:"xpath"`
	Selector   string            `json:"selector"`
	Attributes map[string]string `json:"attributes"`
	Visible    bool              `json:"visible"`
}

func PrintAllPageElementsToFile(page playwright.Page, tcNumber string) {
	elements, err := GetAllElementsStructured(page)
	if err != nil {
		logger.Log("Error getting elements: "+err.Error(), tcNumber)
	}
	jsonData, err := ElementsToJSON(elements)
	if err != nil {
		logger.Log("Error converting elements to JSON: "+err.Error(), tcNumber)
	}
	testingToolkit.PrintStringToFile(jsonData, testRunParameters.GetResultsFolderPath()+"/elements.json")
}

func GetAllElementsStructured(page playwright.Page) ([]Element, error) {

	jsCode := `
		() => {
			function getXPath(element) {
				try {
					if (element.id !== '') {
						return '//*[@id="' + element.id + '"]';
					}
					if (element === document.body) {
						return '/html/body';
					}

					let ix = 0;
					const siblings = element.parentNode.childNodes;
					for (let i = 0; i < siblings.length; i++) {
						const sibling = siblings[i];
						if (sibling === element) {
							return getXPath(element.parentNode) + '/' + element.tagName.toLowerCase() + '[' + (ix + 1) + ']';
						}
						if (sibling.nodeType === 1 && sibling.tagName === element.tagName) {
							ix++;
						}
					}
					return '//unknown';
				} catch(e) {
					return '//error';
				}
			}

			function getCSSSelector(element) {
				try {
					if (element.id) {
						return '#' + element.id;
					}

					const path = [];
					let current = element;

					while (current && current.nodeType === Node.ELEMENT_NODE) {
						let selector = current.nodeName.toLowerCase();

						if (current.classList && current.classList.length > 0) {
							selector += '.' + Array.from(current.classList).join('.');
						}

						if (current.parentNode) {
							const siblings = Array.from(current.parentNode.children);
							const matchingSiblings = siblings.filter(sibling => {
								let siblingSelector = sibling.nodeName.toLowerCase();
								if (sibling.classList && sibling.classList.length > 0) {
									siblingSelector += '.' + Array.from(sibling.classList).join('.');
								}
								return siblingSelector === selector;
							});

							if (matchingSiblings.length > 1) {
								const index = siblings.indexOf(current) + 1;
								selector += ':nth-child(' + index + ')';
							}
						}

						path.unshift(selector);
						current = current.parentNode;

						if (current && (current.nodeName.toLowerCase() === 'body' || current.nodeName.toLowerCase() === 'html')) {
							break;
						}
					}

					return path.join(' > ');
				} catch(e) {
					return 'error-selector';
				}
			}

			function getOptimalSelector(element) {
				try {
					if (element.id) {
						return '#' + element.id;
					}

					const uniqueAttrs = ['data-testid', 'data-test', 'data-cy', 'data-automation-id', 'name'];
					for (const attr of uniqueAttrs) {
						if (element.hasAttribute(attr)) {
							const value = element.getAttribute(attr);
							if (value) {
								return '[' + attr + '="' + value + '"]';
							}
						}
					}

					if (element.classList && element.classList.length > 0) {
						const classSelector = '.' + Array.from(element.classList).join('.');
						const matches = document.querySelectorAll(classSelector);
						if (matches.length === 1) {
							return classSelector;
						}
					}

					return getCSSSelector(element);
				} catch(e) {
					return 'error-optimal-selector';
				}
			}

			function isVisible(element) {
				try {
					const style = window.getComputedStyle(element);
					return style.display !== 'none' &&
						   style.visibility !== 'hidden' &&
						   style.opacity !== '0' &&
						   element.offsetParent !== null;
				} catch(e) {
					return false;
				}
			}

			function getElementAttributes(element) {
				try {
					const attrs = {};
					for (const attr of element.attributes) {
						attrs[attr.name] = attr.value;
					}
					return attrs;
				} catch(e) {
					return {};
				}
			}

			const allElements = Array.from(document.querySelectorAll('*:not(div):not(script):not(style):not(span):not(meta):not(link):not(noscript)'));

			for (let i = allElements.length - 1; i >= 0; i--) {
				const element = allElements[i];
				if (element.tagName.toLowerCase() === 'div') {
					const textContent = (element.textContent || '').trim();
					if (textContent.Length === 0) {
						allElements.splice(i, 1);
					}
				}	
			}

			return allElements.map(el => {
				try {
					return {
						tag: el.tagName.toLowerCase(),
						id: el.id || '',
						text: (el.textContent || '').trim().substring(0, 100),
						xpath: getXPath(el),
						selector: getOptimalSelector(el),
						attributes: getElementAttributes(el),
						visible: isVisible(el)
					};
				} catch(e) {
					return {
						tag: 'error',
						id: '',
						text: 'Error extracting element',
						xpath: '//error',
						selector: 'error',
						attributes: {},
						visible: false
					};
				}
			});
		}
	`

	result, err := page.Evaluate(jsCode)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate JavaScript: %w", err)
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var elements []Element
	err = json.Unmarshal(jsonBytes, &elements)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal elements: %w", err)
	}

	return elements, nil
}

func ElementsToJSON(elements []Element) (string, error) {
	jsonData, err := json.MarshalIndent(elements, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

/*
jsCode := `
		() => {
			function getXPath(element) {
				try {
					if (element.id !== '') {
						return '//*[@id="' + element.id + '"]';
					}
					if (element === document.body) {
						return '/html/body';
					}

					let ix = 0;
					const siblings = element.parentNode.childNodes;
					for (let i = 0; i < siblings.length; i++) {
						const sibling = siblings[i];
						if (sibling === element) {
							return getXPath(element.parentNode) + '/' + element.tagName.toLowerCase() + '[' + (ix + 1) + ']';
						}
						if (sibling.nodeType === 1 && sibling.tagName === element.tagName) {
							ix++;
						}
					}
					return '//unknown';
				} catch(e) {
					return '//error';
				}
			}

			function getCSSSelector(element) {
				try {
					if (element.id) {
						return '#' + element.id;
					}

					const path = [];
					let current = element;

					while (current && current.nodeType === Node.ELEMENT_NODE) {
						let selector = current.nodeName.toLowerCase();

						if (current.classList && current.classList.length > 0) {
							selector += '.' + Array.from(current.classList).join('.');
						}

						if (current.parentNode) {
							const siblings = Array.from(current.parentNode.children);
							const matchingSiblings = siblings.filter(sibling => {
								let siblingSelector = sibling.nodeName.toLowerCase();
								if (sibling.classList && sibling.classList.length > 0) {
									siblingSelector += '.' + Array.from(sibling.classList).join('.');
								}
								return siblingSelector === selector;
							});

							if (matchingSiblings.length > 1) {
								const index = siblings.indexOf(current) + 1;
								selector += ':nth-child(' + index + ')';
							}
						}

						path.unshift(selector);
						current = current.parentNode;

						if (current && (current.nodeName.toLowerCase() === 'body' || current.nodeName.toLowerCase() === 'html')) {
							break;
						}
					}

					return path.join(' > ');
				} catch(e) {
					return 'error-selector';
				}
			}

			function getOptimalSelector(element) {
				try {
					if (element.id) {
						return '#' + element.id;
					}

					const uniqueAttrs = ['data-testid', 'data-test', 'data-cy', 'data-automation-id', 'name'];
					for (const attr of uniqueAttrs) {
						if (element.hasAttribute(attr)) {
							const value = element.getAttribute(attr);
							if (value) {
								return '[' + attr + '="' + value + '"]';
							}
						}
					}

					if (element.classList && element.classList.length > 0) {
						const classSelector = '.' + Array.from(element.classList).join('.');
						const matches = document.querySelectorAll(classSelector);
						if (matches.length === 1) {
							return classSelector;
						}
					}

					return getCSSSelector(element);
				} catch(e) {
					return 'error-optimal-selector';
				}
			}

			function isVisible(element) {
				try {
					const style = window.getComputedStyle(element);
					return style.display !== 'none' &&
						   style.visibility !== 'hidden' &&
						   style.opacity !== '0' &&
						   element.offsetParent !== null;
				} catch(e) {
					return false;
				}
			}

			function getElementAttributes(element) {
				try {
					const attrs = {};
					for (const attr of element.attributes) {
						attrs[attr.name] = attr.value;
					}
					return attrs;
				} catch(e) {
					return {};
				}
			}

			function divHasText(element) {
				if (element.tagName.toLowerCase() !== 'div') return false;
				const textContent = element.textContent?.trim() || } '';
				return textContent.length > 0;
			}

			// Extract all elements
			const allElements = Array.from(document.querySelectorAll('*:not(script):not(style):not(span):not(meta):not(img):not(link):not(noscript)')).filter(el => el.tagName.toLowerCase() !== 'div' || divHasText(el));

			return allElements.map(el => {
				try {
					return {
						tag: el.tagName.toLowerCase(),
						id: el.id || '',
						text: (el.textContent || '').trim().substring(0, 100),
						xpath: getXPath(el),
						selector: getOptimalSelector(el),
						attributes: getElementAttributes(el),
						visible: isVisible(el)
					};
				} catch(e) {
					return {
						tag: 'error',
						id: '',
						text: 'Error extracting element',
						xpath: '//error',
						selector: 'error',
						attributes: {},
						visible: false
					};
				}
			});
		}
	`
*/
