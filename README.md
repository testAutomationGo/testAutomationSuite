# BUILD UI
To Build The Testing app on Windows Run: go build -ldflags "-H=windowsgui" -o TestingApp.exe ./app/appMain.go

To Build The Testing app on Mac Run: go build -o TestingApp ./app/appMain.go

# CONFIGURATION
In config folder:

Edit appConfig.json for your needs

Edit argsAfterEnv.json for the Testing App arguments if necessary.

Add and edit your env config files. Currently a local, develop, and production file are present add any as necessary though ensure
environments are edited in the appConfig.json for these environments.
