if \
    aws ecr-public describe-images \
        --registry-id $AWS_REGISTRY_ID \
        --repository-name $NOLUS_BUILDER_REPO --region us-east-1 \
        --image-ids=imageTag=$NOLUS_BUILDER_TAG \
then \
    echo "NOLUS_BUILDER_EXISTS=true" >>builder_exists.env \
fi
