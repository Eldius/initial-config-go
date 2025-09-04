
ifeq ($(VERSION),)
     VERSION:=$(shell git describe --tags --abbrev=0 | awk -F .   '{OFS="."; $$NF+=1; print}')
endif


test:
	@echo ""
	@echo "######################"
	@echo "#  Running tests...  #"
	@echo "######################"
	@echo ""
	@echo ""
	go test ./... -cover
	@echo "----------------------"
	@echo ""

lint:
	@echo ""
	@echo "########################"
	@echo "#  Static analyses...  #"
	@echo "########################"
	@echo ""
	@echo ""
	golangci-lint run
	@echo "------------------------"
	@echo ""

vulncheck:
	@echo ""
	@echo "############################"
	@echo "#  Vulnerability check...  #"
	@echo "############################"
	@echo ""
	@echo ""
	govulncheck ./...
	@echo "----------------------------"
	@echo ""

validate: test lint vulncheck
	@echo ""
	@echo "#############################"
	@echo "#  Validation completed...  #"
	@echo "#############################"
	@echo ""
	@echo ""
	@echo "-----------------------------"
	@echo ""

release: test lint vulncheck
	@echo ""
	@echo "##############################"
	@echo "# Generating next version... #"
	@echo "##############################"
	@echo ""
	@echo "Next version: $(VERSION)"
	@echo ""
	@echo ""

	git tag $(VERSION)
	git push
	git push --tags
	@echo "------------------------------"
	@echo ""

benchmark:
	go test \
		-bench=. \
		-benchmem \
		-count=20 \
			./...

telemetry-example-down:
	docker compose -f docker-compose-telemetry.yml down

telemetry-example: telemetry-example-down
	COMPOSE_BAKE=true docker compose -f docker-compose-telemetry.yml up --build
