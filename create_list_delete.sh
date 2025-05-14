#!/bin/bash

# Script to create and delete instances in Digital Ocean via API
# Usage: ./create_list_delete.sh -k <api_key> -h <host_ip> -p <project_name> [-n <number_of_instances>] [-c] [-d]
# Example: ./create_list_delete.sh -k <api-token> -h <host-ip> -p <project-name> -n 2 -c
# Example: ./create_list_delete.sh -k <api-token> -h <host-ip> -p <project-name> -d
# Example: ./create_list_delete.sh -k <api-token> -h <host-ip> -p <project-name> -n 3 -c -d

# Initialize variables
CREATE=false
DELETE=false
API_KEY=""
HOST_IP=""
PROJECT_NAME=""
NUM_INSTANCES=1  # Default to 1 instance

# Parse command line options
while getopts "k:h:p:n:cd" opt; do
    case $opt in
        k)
            API_KEY="$OPTARG"
            ;;
        h)
            HOST_IP="$OPTARG"
            ;;
        p)
            PROJECT_NAME="$OPTARG"
            ;;
        n)
            NUM_INSTANCES="$OPTARG"
            # Validate that NUM_INSTANCES is a positive integer
            if ! [[ "$NUM_INSTANCES" =~ ^[1-9][0-9]*$ ]]; then
                echo "Error: Number of instances must be a positive integer"
                exit 1
            fi
            ;;
        c)
            CREATE=true
            ;;
        d)
            DELETE=true
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            echo "Usage: $0 -k <api_key> -h <host_ip> -p <project_name> [-n <number_of_instances>] [-c] [-d]"
            exit 1
            ;;
        :)
            echo "Option -$OPTARG requires an argument." >&2
            echo "Usage: $0 -k <api_key> -h <host_ip> -p <project_name> [-n <number_of_instances>] [-c] [-d]"
            exit 1
            ;;
    esac
done

# Check if API key is provided
if [ -z "$API_KEY" ]; then
    echo "Error: API key is required"
    echo "Usage: $0 -k <api_key> -h <host_ip> -p <project_name> [-n <number_of_instances>] [-c] [-d]"
    exit 1
fi

# Check if host IP is provided
if [ -z "$HOST_IP" ]; then
    echo "Error: Host IP is required"
    echo "Usage: $0 -k <api_key> -h <host_ip> -p <project_name> [-n <number_of_instances>] [-c] [-d]"
    exit 1
fi

# Check if project name is provided
if [ -z "$PROJECT_NAME" ]; then
    echo "Error: Project name is required"
    echo "Usage: $0 -k <api_key> -h <host_ip> -p <project_name> [-n <number_of_instances>] [-c] [-d]"
    exit 1
fi

# Check if at least one action is specified
if [ "$CREATE" = false ] && [ "$DELETE" = false ]; then
    echo "Error: At least one action (-c or -d) must be specified"
    echo "Usage: $0 -k <api_key> -h <host_ip> -p <project_name> [-n <number_of_instances>] [-c] [-d]"
    exit 1
fi

API_URL="http://${HOST_IP}:8000/talis/api/v1/instances"
PROJECT_URL="http://${HOST_IP}:8000/talis/api/v1"

# Function to create project
create_project() {
    echo "Attempting to create project: $PROJECT_NAME"
    
    PROJECT_PAYLOAD="{
        \"method\": \"project.create\",
        \"params\": {
            \"name\": \"$PROJECT_NAME\",
            \"description\": \"$PROJECT_NAME\",
            \"config\": \"{}\",
            \"owner_id\": 1
        }
    }"
    
    echo "Sending project creation request with payload:"
    echo "$PROJECT_PAYLOAD"
    
    RESPONSE=$(curl -X POST \
         -H "Content-Type: application/json" \
         -H "apikey: $API_KEY" \
         -d "$PROJECT_PAYLOAD" \
         "$PROJECT_URL")
    
    if [ $? -eq 0 ]; then
        echo "Project creation request sent. Server response:"
        echo "$RESPONSE"
    else
        echo "Failed to create project."
        exit 1
    fi
}

# Function to create instances
create_instances() {
    echo "Attempting to create $NUM_INSTANCES instances..."
    
    PAYLOAD_TEMPLATE='''[
    {
        "owner_id": 1,
        "provider": "do",
        "number_of_instances": 1,
        "provision": true,
        "region": "nyc3",
        "size": "s-1vcpu-1gb",
        "image": "ubuntu-22-04-x64",
        "tags": ["talis", "dev", "testing"],
        "ssh_key_name": "talis-dev-server",
        "project_name": "%s",
        "volumes": [
            {
                "name": "talis-volume",
                "size_gb": 20,
                "mount_point": "/mnt/data"
            }
        ]
    }
]'''

    # Create an array to store instance names for deletion
    INSTANCE_NAMES=()
    
    for i in $(seq 0 $((NUM_INSTANCES-1))); do
        CURRENT_PAYLOAD=$(printf "$PAYLOAD_TEMPLATE" "$i" "$PROJECT_NAME")
        INSTANCE_NAME="test-$i"
        INSTANCE_NAMES+=("$INSTANCE_NAME")
        
        echo "Creating instance $INSTANCE_NAME with payload:"
        echo "$CURRENT_PAYLOAD"
        
        RESPONSE=$(curl -X POST \
             -H "Content-Type: application/json" \
             -H "apikey: $API_KEY" \
             -d "$CURRENT_PAYLOAD" \
             "$API_URL")
        
        if [ $? -eq 0 ]; then
            echo "Instance $INSTANCE_NAME creation request sent. Server response:"
            echo "$RESPONSE"
        else
            echo "Failed to send request for instance $INSTANCE_NAME."
        fi
        
        echo "------------------------------------"
        sleep 2
    done
    
    echo "Finished sending requests for $NUM_INSTANCES instances."
}

# Function to delete instances
delete_instances() {
    echo "Attempting to delete instances..."
    
    # Create the instance names array for deletion
    INSTANCE_NAMES=()
    for i in $(seq 0 $((NUM_INSTANCES-1))); do
        INSTANCE_NAMES+=("test-$i")
    done
    
    # Convert array to JSON array string
    INSTANCE_NAMES_JSON=$(printf '"%s",' "${INSTANCE_NAMES[@]}" | sed 's/,$//')
    
    DELETE_PAYLOAD="{
        \"project_name\": \"$PROJECT_NAME\",
        \"instance_ids\": [2] # TODO: Change to use instance names
    }"
    
    echo "Sending delete request with payload:"
    echo "$DELETE_PAYLOAD"
    
    RESPONSE=$(curl -X DELETE \
         -H "Content-Type: application/json" \
         -H "apikey: $API_KEY" \
         -d "$DELETE_PAYLOAD" \
         "$API_URL")
    
    if [ $? -eq 0 ]; then
        echo "Delete request sent. Server response:"
        echo "$RESPONSE"
    else
        echo "Failed to send delete request."
    fi
}

# Execute requested actions
if [ "$CREATE" = true ]; then
    create_project
    create_instances
fi

if [ "$DELETE" = true ]; then
    delete_instances
fi
