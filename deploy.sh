HOST=ubuntu@ec2-54-171-95-164.eu-west-1.compute.amazonaws.com

go build

scp -i "~/.ssh/minekeypair.pem" gangplankbot $HOST:_gangplankbot
scp -i "~/.ssh/minekeypair.pem" fbaccountkey.json $HOST:fbaccountkey.json
scp -i "~/.ssh/minekeypair.pem" runbot.sh $HOST:runbot.sh
scp -i "~/.ssh/minekeypair.pem" riotapi/champions.json $HOST:riotapi/champions.json
scp -i "~/.ssh/minekeypair.pem" riotapi/runesreforged.json $HOST:riotapi/runesreforged.json

ssh -i "~/.ssh/minekeypair.pem" $HOST bash -c "'
./runbot.sh $1 $2
'"

