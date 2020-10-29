# Azure MySQL WordPress Example

1. porter build
1. porter credentials generate
    You will need an Azure service principal. Put the service principal credentials in environment variables.
1. porter install --cred azure-wordpress --param-file params.ini