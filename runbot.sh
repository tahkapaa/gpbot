pkill gangplankbot
# if ! ( screen -ls | grep gbbot > /dev/null); then
#     screen -dmS gbbot;
# fi
screen -X -S gbbot quit
mv -f _gangplankbot gangplankbot
screen -dmS gbbot ./gangplankbot -t $1 -r eune -a $2