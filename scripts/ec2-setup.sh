GROUP_ID=$(aws ec2 describe-security-groups --group-names default | jq '.SecurityGroups | .[].GroupId')
AWS_AMI='/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-ebs'

AWS_AMI_ID="ami-07508eaa6a8ced899"

aws ec2 run-instances --image-id "$AWS_AMI_ID" --count 1 --instance-type t3.micro --key-name gitlab --security-group-ids "$GROUP_ID" --block-device-mappings "[{\"DeviceName\":\"/dev/sdf\",\"Ebs\":{\"VolumeSize\":20,\"DeleteOnTermination\":false}}]"


ssh  -i ./gitlab.pem ec2-user@ec2-35-158-128-53.eu-central-1.compute.amazonaws.com  'sudo yum update -y'
ssh  -i ./gitlab.pem ec2-user@ec2-35-158-128-53.eu-central-1.compute.amazonaws.com  'sudo yum install docker -y'
ssh  -i ./gitlab.pem ec2-user@ec2-35-158-128-53.eu-central-1.compute.amazonaws.com  'sudo service docker start'
ssh  -i ./gitlab.pem ec2-user@ec2-35-158-128-53.eu-central-1.compute.amazonaws.com  'sudo usermod -a -G docker ec2-user'