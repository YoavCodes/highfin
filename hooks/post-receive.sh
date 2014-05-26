# TODO: change to #!/usr/bin/env node

#!/bin/bash
while read oldrev newrev refname
do
    branch=$(git rev-parse --symbolic --abbrev-ref $refname)
    curl http://10.10.10.5/$oldrev/$newrev/$refname/$branch
    if [ "dev-next" == "$branch" ]; then
        # Do something
        curl http://10.10.10.5/$oldrev/$newrev/$refname
    fi
done