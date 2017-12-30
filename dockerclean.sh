#! /bin/bash

echo "Cleaning docker images: $1 | $2"
IFS=$'\n'
images=`docker images -a --format '{{.Repository}},{{.ID}},{{.CreatedSince}}' | grep "none\|$1"`
for i in $images; do
    image=`echo $i | cut -d ',' -f 2`
    time=`echo $i | cut -d ',' -f 3`
    number=`echo $time | cut -d ' ' -f 1`
    unit=`echo $time | cut -d ' ' -f 2`
    if [ "$unit" == "days" ] || [ "$unit" == "day" ]; then
        echo "REMOVE" $image":" $time
        docker rmi -f $image
    elif [ "$unit" == "hours" ] && [ "$number" -ge "$2" ]; then
        echo "REMOVE" $image":" $time
        docker rmi -f $image
    fi
done
