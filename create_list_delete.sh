#!/bin/bash

# Script to create, list, and delete instances in Digital Ocean via API
# Usage: ./create_list_delete.sh -k <api_key> -h <host_ip> -p <project_name> [-n <number_of_instances>] [-c] [-d] [-l]
# Example: ./create_list_delete.sh -k <api-token> -h <host-ip> -p <project-name> -l
# Example: ./create_list_delete.sh -k <api-token> -h <host-ip> -p <project-name> -n 2 -c
# Example: ./create_list_delete.sh -k <api-token> -h <host-ip> -p <project-name> -d
# Example: ./create_list_delete.sh -k <api-token> -h <host-ip> -p <project-name> -n 3 -c -d

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed. Please install jq first."
    echo "You can install it with: brew install jq (macOS) or apt-get install jq (Ubuntu)"
    exit 1
fi

# Initialize variables
CREATE=false
DELETE=false
LIST=false
API_KEY=""
HOST_IP=""
PROJECT_NAME=""
NUM_INSTANCES=1  # Default to 1 instance

# Parse command line options
while getopts "k:h:p:n:cld" opt; do
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
        l)
            LIST=true
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            echo "Usage: $0 -k <api_key> -h <host_ip> -p <project_name> [-n <number_of_instances>] [-c] [-d] [-l]"
            exit 1
            ;;
        :)
            echo "Option -$OPTARG requires an argument." >&2
            echo "Usage: $0 -k <api_key> -h <host_ip> -p <project_name> [-n <number_of_instances>] [-c] [-d] [-l]"
            exit 1
            ;;
    esac
done

# Check if API key is provided
if [ -z "$API_KEY" ]; then
    echo "Error: API key is required"
    echo "Usage: $0 -k <api_key> -h <host_ip> -p <project_name> [-n <number_of_instances>] [-c] [-d] [-l]"
    exit 1
fi

# Check if host IP is provided
if [ -z "$HOST_IP" ]; then
    echo "Error: Host IP is required"
    echo "Usage: $0 -k <api_key> -h <host_ip> -p <project_name> [-n <number_of_instances>] [-c] [-d] [-l]"
    exit 1
fi

# Check if project name is provided
if [ -z "$PROJECT_NAME" ]; then
    echo "Error: Project name is required"
    echo "Usage: $0 -k <api_key> -h <host_ip> -p <project_name> [-n <number_of_instances>] [-c] [-d] [-l]"
    exit 1
fi

# Check if at least one action is specified
if [ "$CREATE" = false ] && [ "$DELETE" = false ] && [ "$LIST" = false ]; then
    echo "Error: At least one action (-c, -d, or -l) must be specified"
    echo "Usage: $0 -k <api_key> -h <host_ip> -p <project_name> [-n <number_of_instances>] [-c] [-d] [-l]"
    exit 1
fi

API_URL="http://${HOST_IP}:8000/talis/api/v1/instances"
PROJECT_URL="http://${HOST_IP}:8000/talis/api/v1"
LIST_URL="http://${HOST_IP}:8000/talis/api/v1/instances/all-metadata"

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
    
    PAYLOAD_TEMPLATE='[
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
]'

    # Create an array to store instance IDs for deletion
    INSTANCE_IDS=()
    
    for i in $(seq 1 $NUM_INSTANCES); do
        CURRENT_PAYLOAD=$(printf "$PAYLOAD_TEMPLATE" "$PROJECT_NAME")
        
        echo "Creating instance $i with payload:"
        echo "$CURRENT_PAYLOAD"
        
        # Make the request and capture both response and status code
        RESPONSE=$(curl -s -w "\n%{http_code}" \
             -H "Content-Type: application/json" \
             -H "apikey: $API_KEY" \
             -d "$CURRENT_PAYLOAD" \
             "$API_URL")
        
        # Extract the status code (last line) and response body
        HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
        RESPONSE_BODY=$(echo "$RESPONSE" | sed '$d')
        
        if [ "$HTTP_CODE" -eq 200 ] || [ "$HTTP_CODE" -eq 201 ]; then
            echo "Instance creation request sent. Server response:"
            echo "$RESPONSE_BODY"
            
            # Extract instance ID from response using jq
            INSTANCE_ID=$(echo "$RESPONSE_BODY" | jq -r '.result.id')
            if [ "$INSTANCE_ID" != "null" ] && [ ! -z "$INSTANCE_ID" ]; then
                echo "Successfully created instance with ID: $INSTANCE_ID"
                INSTANCE_IDS+=("$INSTANCE_ID")
            else
                echo "Warning: Could not extract instance ID from response"
            fi
        else
            echo "Failed to create instance. HTTP Status: $HTTP_CODE"
            echo "Response: $RESPONSE_BODY"
        fi
        
        echo "------------------------------------"
        sleep 2
    done
    
    # Export the instance IDs for use in delete function
    export INSTANCE_IDS
    
    echo "Finished creating $NUM_INSTANCES instances."
    echo "Created instance IDs: ${INSTANCE_IDS[*]}"
}

# Function to delete instances
delete_instances() {
    echo "Attempting to delete instances..."
    
    if [ ${#INSTANCE_IDS[@]} -eq 0 ]; then
        echo "No instance IDs available for deletion"
        return
    fi
    
    # Convert array to JSON array string
    INSTANCE_IDS_JSON=4
    #INSTANCE_IDS_JSON=$(printf '%s,' "${INSTANCE_IDS[@]}" | sed 's/,$//')
    
    DELETE_PAYLOAD="{
        \"project_name\": \"$PROJECT_NAME\",
        \"instance_ids\": [$INSTANCE_IDS_JSON]
    }"
    
    echo "Sending delete request with payload:"
    echo "$DELETE_PAYLOAD"
    
    # Make the request and capture both response and status code
    RESPONSE=$(curl -s -w "\n%{http_code}" \
         -X DELETE \
         -H "Content-Type: application/json" \
         -H "apikey: $API_KEY" \
         -d "$DELETE_PAYLOAD" \
         "$API_URL")
    
    # Extract the status code (last line) and response body
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    RESPONSE_BODY=$(echo "$RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" -eq 200 ] || [ "$HTTP_CODE" -eq 204 ]; then
        echo "Successfully deleted instances. HTTP Status: $HTTP_CODE"
        echo "Response: $RESPONSE_BODY"
    else
        echo "Failed to delete instances. HTTP Status: $HTTP_CODE"
        echo "Response: $RESPONSE_BODY"
    fi
}

# Function to list instances
list_instances() {
    echo "Listing instances for project: $PROJECT_NAME"
    
    # Make the request and capture both response and status code
    RESPONSE=$(curl -s -w "\n%{http_code}" \
         -H "Content-Type: application/json" \
         -H "apikey: $API_KEY" \
         "$LIST_URL")
    
    # Extract the status code (last line) and response body
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    RESPONSE_BODY=$(echo "$RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" -eq 200 ]; then
        echo "Successfully retrieved instances. HTTP Status: $HTTP_CODE"
        # Use jq to format the output nicely
        echo "$RESPONSE_BODY" | jq '.'
    else
        echo "Failed to list instances. HTTP Status: $HTTP_CODE"
        echo "Response: $RESPONSE_BODY"
    fi
}

# Execute requested actions
if [ "$LIST" = true ]; then
    list_instances
fi

if [ "$CREATE" = true ]; then
    create_project
    create_instances
fi

if [ "$DELETE" = true ]; then
    delete_instances
fi
