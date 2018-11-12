For Linux based systems, install the MongoDB into your distribution
and start the database. For Fedora, instructions are contained here:

https://developer.fedoraproject.org/tech/database/mongodb/about.html

However it essentially renders down to (on Fedora)

dnf install mongodb mongodb-server
service mongod start

Ensure your system has a standard GO installation.

You'll also need a few other additional open source products to run
this program. Here are the various go get instructions:

go get gopkg.in/mgo.v2
go get github.com/smartystreets/goconvey
go get github.com/gorilla/mux

Build this project with a simple "go build" command.

Tests are run with a simple "go test -v" command.

You can view the output of the tests in graphical format by running:

$GOPATH/bin/goconvey
