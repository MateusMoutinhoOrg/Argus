add a description on callback to be ploted on help menssage

~~~go 
	props := Argus.GenerationProps{
		Callbacks: []Argus.Callback{
			{Starter: "serve", Callback: serve, Description: "Serve the application"},
			{Starter: "status", Callback: status, Description: "Get the application status"},
		},
	}
~~~