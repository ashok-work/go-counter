package main

func main() {
	runWebViewApp()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
