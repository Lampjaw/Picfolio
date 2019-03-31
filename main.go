package main

func main() {
	exitCallback := make(chan bool)

	repository := InitRepository()

	adminServer := newAdminServer(repository)
	publicServer := newPublicServer(repository)

	adminServer.startListeningAdmin(exitCallback)
	publicServer.startListeningPublic(exitCallback)

	<-exitCallback
}
