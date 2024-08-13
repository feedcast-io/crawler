prod:
	go mod vendor
	gcloud \
		--project=feedcast-2023 \
		functions deploy web-crawler \
		--gen2 \
		--runtime=go122 \
		--region=europe-west9 \
		--trigger-http \
		--allow-unauthenticated \
		--entry-point=crawl \
		--source . \
		--cpu=0.166 \
		--memory=128Mi \
		--timeout=300 \
		--min-instances=0 \
		--max-instances=1 \