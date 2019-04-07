package main

func main() {
	appState := newAppState()

	appState.initRepository()

	adminServer := newAdminServer(appState)
	publicServer := newPublicServer(appState)

	adminServer.startListeningAdmin()
	publicServer.startListeningPublic()

	<-appState.exitCallback
}
