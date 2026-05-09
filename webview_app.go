package main

import (
	"fmt"
	"log"
	"net/url"

	webview "github.com/webview/webview_go"
)

func runWebViewApp() {
	w := webview.New(true)
	defer w.Destroy()

	width, height := initialWindowSize()

	w.SetTitle("Go Desktop Browser")
	w.SetSize(width, height, webview.HintNone)

	bindBridge(w)

	w.Navigate("https://ccuat.tricubeinnosoft.com/counter")
	w.Run()
}

func bindBridge(w webview.WebView) {
	must(w.Bind("goPing", func(message string) map[string]string {
		log.Printf("goPing called with message: %s", message)
		return map[string]string{
			"message": fmt.Sprintf("Go received: %s", message),
		}
	}))

	must(w.Bind("goNavigate", func(rawURL string) error {
		u, err := url.ParseRequestURI(rawURL)
		if err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}

		w.Dispatch(func() {
			w.Navigate(u.String())
		})
		return nil
	}))

	must(w.Bind("goGetTestOrder", func() Order {
		return testOrder
	}))

	must(w.Bind("goGetPrinters", func() (PrintersResponse, error) {
		printers, err := getPrinters()
		if err != nil {
			return PrintersResponse{}, err
		}

		return PrintersResponse{
			Printers:          printers,
			SelectedPrinter:   getSelectedPrinter(),
			SelectedPaperSize: getSelectedPaperSize(),
		}, nil
	}))

	must(w.Bind("goSetPrinter", func(printerName string) bool {
		setSelectedPrinter(printerName)
		return true
	}))

	must(w.Bind("goSelectedPrinter", func() string {
		return getSelectedPrinter()
	}))

	must(w.Bind("goDeletePrinter", func() bool {
		setSelectedPrinter("")
		return true
	}))

	must(w.Bind("goSetPaperSize", func(paperSize string) bool {
		setSelectedPaperSize(paperSize)
		return true
	}))

	must(w.Bind("goTestPrint", func() (bool, error) {
		return true, printSimpleTestPage()
	}))

	must(w.Bind("goTestPrintOrder", func() (bool, error) {
		if err := printReceiptPDF(testOrder); err != nil {
			return false, err
		}
		if testOrder.EnableKOTPrint {
			if err := printKOTPDF(testOrder); err != nil {
				return false, err
			}
		}
		return true, nil
	}))

	must(w.Bind("goPrintOrder", func(request PrintOrderRequest) (bool, error) {
		return true, printReceiptPDF(request.Order)
	}))

	must(w.Bind("goPrintKOTOrder", func(request PrintOrderRequest) (bool, error) {
		return true, printKOTPDF(request.Order)
	}))

	w.Init(`
		window.goApp = {
			ping(message) {
				return window.goPing(message);
			},
			navigate(url) {
				return window.goNavigate(url);
			},
			getTestOrder() {
				return window.goGetTestOrder();
			},
			printTestOrder() {
				return window.goTestPrintOrder();
			},
			printOrder(order) {
				return window.goPrintOrder({ order });
			},
			printKOTOrder(order) {
				return window.goPrintKOTOrder({ order });
			},
		};

		window.electronAPI = {
			printOrder(data) {
				return window.goPrintOrder(data);
			},
			printKOTOrder(data) {
				return window.goPrintKOTOrder(data);
			},
			getPrinters() {
				return window.goGetPrinters();
			},
			setPrinter(printerName) {
				return window.goSetPrinter(printerName);
			},
			getSelectedPrinter() {
				return window.goSelectedPrinter();
			},
			deletePrinter() {
				return window.goDeletePrinter();
			},
			setPaperSize(paperSize) {
				return window.goSetPaperSize(paperSize);
			},
			testPrintOrder() {
				return window.goTestPrintOrder();
			},
			testPrint() {
				return window.goTestPrint();
			},
		};
	`)
}
