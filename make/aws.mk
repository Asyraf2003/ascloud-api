# AWS Ops helpers (interactive-safe)
# - Pin 2 region: hosting (SQS/DDB) vs lambda (Lambda/Logs)
# - Hindari script panjang + heredoc; tiap target pendek & deterministik.
# - Semua var bisa dioverride: `make aws-sqs-list AWS_REGION_HOSTING=...`

.PHONY: aws-whoami aws-sqs-list aws-sqs-attrs aws-dlq-ensure aws-redrive-set aws-alarms-sqs aws-alarms-lambda aws-ddb-find-audit aws-logs-find

AWS ?= aws

AWS_REGION_HOSTING ?= ap-southeast-3
AWS_REGION_LAMBDA  ?= ap-southeast-1

HOSTING_QUEUE_NAME ?= your-api-dev-hosting-deploy
HOSTING_DLQ_NAME   ?= $(HOSTING_QUEUE_NAME)-dlq
MAX_RECEIVE_COUNT  ?= 5

AUDIT_TABLE ?= audit_events
LAMBDA_WORKER ?= AsyrafCloud-Unzipper

AWS3 := $(AWS) --region $(AWS_REGION_HOSTING)
AWS1 := $(AWS) --region $(AWS_REGION_LAMBDA)

aws-whoami:
	@echo "== hosting region: $(AWS_REGION_HOSTING) =="; $(AWS3) sts get-caller-identity
	@echo "== lambda   region: $(AWS_REGION_LAMBDA) ==";  $(AWS1) sts get-caller-identity

aws-sqs-list:
	@$(AWS3) sqs list-queues --query 'QueueUrls' --output table

aws-sqs-attrs:
	@Q_URL="$$( $(AWS3) sqs get-queue-url --queue-name "$(HOSTING_QUEUE_NAME)" --query 'QueueUrl' --output text )"; \
	$(AWS3) sqs get-queue-attributes --queue-url "$$Q_URL" --attribute-names QueueArn RedrivePolicy VisibilityTimeout ReceiveMessageWaitTimeSeconds --output json

aws-dlq-ensure:
	@DLQ_URL="$$( $(AWS3) sqs get-queue-url --queue-name "$(HOSTING_DLQ_NAME)" --query 'QueueUrl' --output text 2>/dev/null || true )"; \
	if [ -z "$$DLQ_URL" ] || [ "$$DLQ_URL" = "None" ]; then \
		echo "DLQ missing -> create $(HOSTING_DLQ_NAME)"; \
		DLQ_URL="$$( $(AWS3) sqs create-queue --queue-name "$(HOSTING_DLQ_NAME)" --query 'QueueUrl' --output text )"; \
	fi; \
	echo "DLQ_URL=$$DLQ_URL"; \
	$(AWS3) sqs get-queue-attributes --queue-url "$$DLQ_URL" --attribute-names QueueArn --output json

aws-redrive-set: aws-dlq-ensure
	@Q_URL="$$( $(AWS3) sqs get-queue-url --queue-name "$(HOSTING_QUEUE_NAME)" --query 'QueueUrl' --output text )"; \
	DLQ_URL="$$( $(AWS3) sqs get-queue-url --queue-name "$(HOSTING_DLQ_NAME)" --query 'QueueUrl' --output text )"; \
	DLQ_ARN="$$( $(AWS3) sqs get-queue-attributes --queue-url "$$DLQ_URL" --attribute-names QueueArn --query 'Attributes.QueueArn' --output text )"; \
	ATTR="$$(printf '{"RedrivePolicy":"{\\"deadLetterTargetArn\\":\\"%s\\",\\"maxReceiveCount\\":\\"%s\\"}"}' "$$DLQ_ARN" "$(MAX_RECEIVE_COUNT)")"; \
	$(AWS3) sqs set-queue-attributes --queue-url "$$Q_URL" --attributes "$$ATTR"; \
	$(AWS3) sqs get-queue-attributes --queue-url "$$Q_URL" --attribute-names QueueArn RedrivePolicy --output json

aws-alarms-sqs:
	@$(AWS3) cloudwatch describe-alarms --alarm-names \
	  "$(HOSTING_QUEUE_NAME)-queue-lag-high" \
	  "$(HOSTING_QUEUE_NAME)-backlog-visible" \
	  "$(HOSTING_QUEUE_NAME)-dlq-not-empty" \
	  --output table

aws-alarms-lambda:
	@$(AWS1) cloudwatch describe-alarms --alarm-names \
	  "$(LAMBDA_WORKER)-errors" \
	  --output table

aws-ddb-find-audit:
	@echo "== DDB tables in $(AWS_REGION_HOSTING) matching 'audit' =="; \
	$(AWS3) dynamodb list-tables --query 'TableNames' --output text | tr '\t' '\n' | grep -i audit || true; \
	echo "== describe $(AUDIT_TABLE) in $(AWS_REGION_HOSTING) =="; \
	$(AWS3) dynamodb describe-table --table-name "$(AUDIT_TABLE)" --query "Table.{Name:TableName,Status:TableStatus}" --output table || true; \
	echo "== describe $(AUDIT_TABLE) in $(AWS_REGION_LAMBDA) =="; \
	$(AWS1) dynamodb describe-table --table-name "$(AUDIT_TABLE)" --query "Table.{Name:TableName,Status:TableStatus}" --output table || true

aws-logs-find:
	@$(AWS1) logs describe-log-groups --query "logGroups[?contains(logGroupName, '$(LAMBDA_WORKER)')].logGroupName" --output text
