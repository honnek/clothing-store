##
## DOCKER
## -----------
up:
	docker-compose up -d

down:
	docker-compose down

clear:
	docker system prune --all --volumes --force

build:
	docker-compose down -v --remove-orphans
	docker-compose rm -vsf
	docker-compose up -d --build


##
## UTILS
## -----------
psql-connect:
	docker-compose exec postgres psql -U postgres ranked_choice_2



##
## TESTING
## -----------
test:
	docker-compose exec php sh ./bin/run-tests.sh


##
## REFACTORING
## -----------
fixer:
	docker-compose exec php vendor/bin/php-cs-fixer fix src --verbose

phpstan:
	docker-compose exec php vendor/bin/phpstan analyse src/



bash:
	docker-compose exec -u www-data php bash

bash_root:
	docker-compose exec -u 0 php bash

