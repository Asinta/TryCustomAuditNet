#!/usr/bin/env sh

if [ $# = 0 ]; then
  echo "
  Usage: ./go [commands]
  
  Commands:
    build                                     Build or rebuild services
    up                                        Start all services
    down                                      Stop and remove services
    restart                                   Stop and remove all services then start them all
    run [name]                                Build and start all or one services
    restore                                   Copies and restores the given backup file to the mssql container. If no backup file specified restores the last one.
    shrink                                    Shrink DB with current processing month data only
    backup                                    Create a DB backup
    reindex                                   Reindex data in ElasticSearch
    sqs create-standard-queue [name]          Create SQS standard queue with name (default job)
    sqs create-fifo-queue [name]              Create SQS fifo queue with name (default sync.fifo)
    sqs receive-message [name]                Receive message from SQS queue with name (default sync.fifo)
    sqs send-message [name]                   Send message to SQS queue
    sqs all-attributes [name]                 Fetch all attributes of SQS queue
    s3 create-bucket [bucket]                 Create S3 bucket with bucket name (default bucket)
    s3 list-objects [bucket]                  List S3 objects using bucket name
    s3 get-object [bucket] [key] [outfile]    Get S3 object using bucket name and key
    s3 put-object [bucket] [key] [infile]     Put S3 object using bucket name and key

  Examples: 
    ./go up
    ./go restore path/to/backup-file.bak
    ./go restore      (Restores the last backup.)

  You can also use multiple commands at once: ./go build up"
fi

export AWS_ACCESS_KEY_ID=awskey
export AWS_SECRET_ACCESS_KEY=awssecret
export AWS_DEFAULT_REGION=ap-southeast-2
export AWS_DEFAULT_OUTPUT=json


SQS_ENDPOINT_URL=${SQS_ENDPOINT_URL-"http://localhost:4576"}
S3_ENDPOINT_URL=${S3_ENDPOINT_URL-"http://localhost:4572"}
DEFAULT_FIFO_QUEUE_NAME="sync.fifo"
DEFAULT_STANDARD_QUEUE_NAME="audit"
DEFAULT_BUCKET_NAME="bucket"

MESSAGE_BODY="{\"payload\":{\"notesIdentifier\":\"FA79DD4BF7273006CA2574C600276A8F\",\"loanAmount\":270749.5,\"balance\":1000.0,\"lenderLoanNumber\":\"1224232\",\"settlementDate\":\"2018-09-01\",\"paymentMonth\":\"2018-05-01\",\"paidOutDate\":\"2018-05-01\"},\"endpoint\":\"UpdateLoan\"}"

MESSAGE_BODY="{\"type\":\"EXPORT_LOAN_COMMISSION\",\"payload\":{\"processingMonthId\":\"00000000-0000-0000-0000-000000000009\"}}"

MESSAGE_BODY="{\"type\":\"CONVERT_LN_LOAN_COMMISSION\",\"payload\":{\"processingMonth\":\"2016-07-01\",\"s3key\":\"commissions.csv\"}}"

MESSAGE_BODY="{\"type\":\"CONVERT_LN_INSURANCE_COMMISSION\",\"payload\":{\"processingMonth\":\"2016-07-01\",\"s3key\":\"insurance-commissions.csv\"}}"

MESSAGE_BODY="{\"type\":\"EXPORT_INVOICE\",\"payload\":{\"processingMonthId\":\"00000000-0000-0000-0000-000000000011\"}}"

MESSAGE_BODY="{\"type\":\"EXPORT_COMMISSION_MANAGEMENT_REPORTS\",\"payload\":{\"processingMonthId\":\"00000000-0000-0000-0000-000000000009\"}}"

create_fifo_queue() {
    trap "exit 1" INT
    until aws --endpoint-url=${SQS_ENDPOINT_URL} sqs get-queue-url --queue-name $1 > /dev/null 2>&1; do
        aws --endpoint-url=${SQS_ENDPOINT_URL} sqs create-queue --queue-name $1 --attributes "FifoQueue=true"
    done
    trap - INT
}

create_standard_queue() {
    trap "exit 1" INT
    until aws --endpoint-url=${SQS_ENDPOINT_URL} sqs get-queue-url --queue-name $1 > /dev/null 2>&1; do
        aws --endpoint-url=${SQS_ENDPOINT_URL} sqs create-queue --queue-name $1
    done
    trap - INT
}

send_message() {
 if [[ $1 == *.fifo ]]; then
    aws --endpoint-url=${SQS_ENDPOINT_URL} sqs send-message \
        --queue-url ${SQS_ENDPOINT_URL}/queue/$1 \
        --message-body "${MESSAGE_BODY}" \
        --message-deduplication-id $(date +"%s") \
        --message-group-id $(uuidgen)
 else
    aws --endpoint-url=${SQS_ENDPOINT_URL} sqs send-message \
        --queue-url ${SQS_ENDPOINT_URL}/queue/$1 \
        --message-body "${MESSAGE_BODY}"
 fi
}

receive_message() {
    aws --endpoint-url=${SQS_ENDPOINT_URL} sqs receive-message --queue-url ${SQS_ENDPOINT_URL}/queue/$1 --attribute-names All
}

all_attributes() {
    aws --endpoint-url=${SQS_ENDPOINT_URL} sqs get-queue-attributes --queue-url ${SQS_ENDPOINT_URL}/queue/$1 --attribute-names All
}

create_bucket() {
    trap "exit 1" INT
    until aws --endpoint-url=${S3_ENDPOINT_URL} s3api head-bucket --bucket $1 > /dev/null 2>&1; do
        aws --endpoint-url=${S3_ENDPOINT_URL} s3api create-bucket --bucket $1
    done
    trap - INT
}

list_objects() {
    aws --endpoint-url=${S3_ENDPOINT_URL} s3api list-objects --bucket $1
}

get_object() {
    aws --endpoint-url=${S3_ENDPOINT_URL} s3api get-object --bucket $1 --key $2 $3
}

put_object() {
    aws --endpoint-url=${S3_ENDPOINT_URL} s3api put-object --bucket $1 --key $2 --body $3
}

restart_lambda() {
    docker-compose run --rm create-queues
    docker-compose restart lambda
}

get_mssql_container() {
    mssql_container=$(docker-compose ps -q mssql)
    if [ -z "$mssql_container" ]; then
        echo "Sql container is not running. Starting the container..."
        docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d mssql
        mssql_container=$(docker-compose ps -q mssql)
    fi
}

if [ "$1" = "build" -o "$2" = "build" ]; then
  docker-compose -f docker-compose.yml -f docker-compose.dev.yml build
fi

if [ "$1" = "run" ]; then
  docker-compose rm -f -s $2
  docker-compose -f docker-compose.yml up --build -d $2
  docker-compose logs -f create-queues-audit
fi

if [ "$1" = "down" ]; then
  docker-compose down
fi

if [ "$1" = "restart" ]; then
  docker-compose -f docker-compose.yml -f docker-compose.dev.yml restart
  restart_lambda
  docker-compose logs -f web
fi

if [ "$1" = "up" -o "$2" = "up" ]; then
  docker-compose -f docker-compose.yml up -d
fi

if [ "$1" = "restore" ]; then
    get_mssql_container

    if [ $# -lt 2 ]; then
        echo "No backup file specified to copy. Last backup file will be restored."
    else
        backup=$2
        docker exec ${mssql_container} mkdir -p /var/opt/mssql/backups
        echo "Copying the backup file to the container..."
        docker cp ${backup} ${mssql_container}:/var/opt/mssql/backups/last.bak
    fi

    query="USE [master]
    GO
    ALTER DATABASE [SmartlineConnect] SET SINGLE_USER WITH ROLLBACK IMMEDIATE
    GO
    RESTORE DATABASE [SmartlineConnect] FROM  DISK = N'/var/opt/mssql/backups/last.bak' WITH  FILE = 1,  MOVE N'SmartlineConnect' TO N'/var/opt/mssql/data/SmartlineConnect.mdf',  MOVE N'SmartlineConnect_log' TO N'/var/opt/mssql/data/SmartlineConnect_log.ldf',  NOUNLOAD,  REPLACE,  STATS = 5
    GO
    ALTER DATABASE [SmartlineConnect] SET MULTI_USER
    GO"

    echo "Restoring backup..."
    password=$(docker exec ${mssql_container} bash -c 'echo $SA_PASSWORD')
    docker exec -t ${mssql_container} opt/mssql-tools/bin/sqlcmd -U sa -P ${password} -e -Q "$query"
fi

if [ "$1" = "backup" ]; then
    get_mssql_container

    if [ $# -lt 2 ]; then
        backup="SmartlineConnect.bak"
        echo "No backup file path specified. Will export to ${backup}"
    else
        backup=$2
    fi

    docker exec ${mssql_container} mkdir -p /var/opt/mssql/backups

    query="BACKUP DATABASE [SmartlineConnect] TO DISK = N'/var/opt/mssql/backups/SmartlineConnect.bak' WITH FORMAT, INIT"

    password=$(docker exec ${mssql_container} bash -c 'echo $SA_PASSWORD')
    docker exec -t ${mssql_container} opt/mssql-tools/bin/sqlcmd -U sa -P ${password} -e -Q "$query"

    echo "Copying the backup file from the container..."
    docker cp ${mssql_container}:/var/opt/mssql/backups/SmartlineConnect.bak ${backup}
fi

if [ "$1" = "shrink" ]; then
    get_mssql_container

    # large query truncated when pass to docker exec, so split the query

    query1="USE [SmartlineConnect]
    GO

    DELETE lc
    FROM Commissions.LoanCommission lc
    INNER JOIN Commissions.ProcessingMonth pm ON lc.ReceivedProcessingMonthId = pm.Id
    WHERE pm.IsCurrent = 0
    GO

    DELETE a
    FROM Commissions.Adjustment a
    INNER JOIN Commissions.ProcessingMonth pm ON a.ProcessingMonthId = pm.Id
    WHERE pm.IsCurrent = 0
    GO

    DELETE ic
    FROM Commissions.InsuranceCommission ic
    INNER JOIN Commissions.ProcessingMonth pm ON ic.ProcessingMonthId = pm.Id
    WHERE pm.IsCurrent = 0
    GO

    DELETE lp
    FROM Commissions.LenderPayment lp
    INNER JOIN Commissions.ProcessingMonth pm ON lp.ProcessingMonthId = pm.Id
    WHERE pm.IsCurrent = 0
    GO

    DELETE po
    FROM Commissions.PaidOuts po
    INNER JOIN Commissions.ProcessingMonth pm ON po.ProcessingMonthId = pm.Id
    WHERE pm.IsCurrent = 0
    GO

    DELETE ti
    FROM Commissions.TaxInvoice ti
    INNER JOIN Commissions.ProcessingMonth pm ON ti.ProcessingMonthId = pm.Id
    WHERE pm.IsCurrent = 0
    GO"

    query2="USE [SmartlineConnect]
    GO

    TRUNCATE TABLE NotesMigrationHistory
    GO"

    password=$(docker exec ${mssql_container} bash -c 'echo $SA_PASSWORD')
    docker exec -t ${mssql_container} opt/mssql-tools/bin/sqlcmd -U sa -P ${password} -e -Q "$query1"
    docker exec -t ${mssql_container} opt/mssql-tools/bin/sqlcmd -U sa -P ${password} -e -Q "$query2"
fi

if [ "$1" = "reindex" ]; then
  docker-compose stop logstash
  docker-compose run --rm logstash indexes.sh delete
  docker-compose up --build -d logstash
  docker-compose logs -f logstash
fi

if [ "$1" = "sqs" ]; then
    if [ "$2" = "create-fifo-queue" ]; then
        create_fifo_queue ${3-${DEFAULT_FIFO_QUEUE_NAME}}
    fi

    if [ "$2" = "create-standard-queue" ]; then
        create_standard_queue ${3-${DEFAULT_STANDARD_QUEUE_NAME}}
    fi

    if [ "$2" = "receive-message" ]; then
        receive_message ${3-${DEFAULT_FIFO_QUEUE_NAME}}
    fi
    
    if [ "$2" = "send-message" ]; then
        send_message ${3-${DEFAULT_FIFO_QUEUE_NAME}}
    fi
    
    if [ "$2" = "all-attributes" ]; then
         all_attributes ${3-${DEFAULT_FIFO_QUEUE_NAME}}
    fi
fi

if [ "$1" = "s3" ]; then
    if [ "$2" = "create-bucket" ]; then
        create_bucket ${3-${DEFAULT_BUCKET_NAME}}
    fi

    if [ "$2" = "list-objects" ]; then
        list_objects ${3-${DEFAULT_BUCKET_NAME}}
    fi

    if [ "$2" = "get-object" ]; then
        get_object $3 $4 $5
    fi

    if [ "$2" = "put-object" ]; then
        put_object $3 $4 $5
    fi
fi
