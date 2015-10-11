# := means run the commands once
npmbin := $(shell npm bin)
gobin := $(GOPATH)/bin

all: clean sass js bindata.go

debug: clean sass jsdebug bindata.go

sass:
	@$(npmbin)/node-sass --output-style compressed styles/app.sass app.css

# vendored highlight.js file created with the following command:
# node tools/build.js -n apache bash coffeescript cpp cs css diff dockerfile \
#   erlang go haskell http ini java javascript json less lisp lua makefile xml \
#   markdown nginx objectivec perl php protobuf python r ruby rust scala scss \
#   smalltalk sql stylus swift thrift
js:
	@$(npmbin)/browserify app/app.js \
	  -t babelify \
	  -g uglifyify \
	  --outfile app.js
	@echo "Wrote app.js"

jsdebug:
	@$(npmbin)/browserify app/app.js \
	  -t babelify \
	  --debug \
	  --outfile app.js
	@echo "Wrote app.js with map"

bindata.go:
	@$(gobin)/go-bindata \
	  -ignore "\.DS_Store" \
	  -pkg "static" \
	  app.js \
	  app.css \
	  vendor \
	  compare.html
	@echo "Wrote bindata.go"

clean:
	@[ ! -e bindata.go ] || rm bindata.go
	@[ ! -e app.css ] || rm app.css
	@[ ! -e app.js ] || rm app.js

