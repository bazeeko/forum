    #!/bin/bash
    docker build --tag forum .
    docker run --publish 8070:8070 --detach --name forumt forum