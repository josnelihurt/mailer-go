
all:
	docker build . 

push-docker:
	docker build . -t josnelihurt/mailer-go
	docker push josnelihurt/mailer-go