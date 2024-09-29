package rest

func ExampleDebugAccessCallback() {
	err := StartSimpleService(":8080", "/", "123", DebugAccessCallback)
	if err != nil {
		// handle error
	}
}

func ExampleFullAccessCallback() {
	err := StartSimpleService(":8080", "/", "123", FullAccessCallback)
	if err != nil {
		// handle error
	}
}
