# Custom Windows CI Agent

We maintain our own custom virtual machine image in Azure that we use in Azure Pipelines to run tests on a Windows machine.
The default Windows agents provided by Azure Pipelines are not legally allowed to have Docker Desktop installed, which is why we need to maintain our own agent image.

These are the components that are used to manage our Windows agent:
* A virtual machine scale set hosted that is registered with Azure Pipelines so that new agents can be created as needed.
* A generalized image from which new virtual machines can be created.
  The scale set is configured to create new agents using this image.
* The original virtual disk that the generalized image was created from. 
  We keep this around so that it's easier to update the agent image over time, without having to repeat the entire vm setup.
  Don't keep the entire vm around, that's much more expensive.

## Configuration
The image has the following additional software:

* WSL2
* Docker Desktop for Windows
* Azure Pipelines Agent

It is configured to automatically start Docker Desktop when the agent starts (normally Docker Desktop only starts when a user logs in).
The service is called "Docker Desktop" and is equivalent to double-clicking on the Docker Desktop icon, and it runs as our CI user, porterci.
There is another service that is "Docker Desktop Service", but that only runs the backend and isn't sufficient for commands like `docker ps` to work.
Both are needed.

The agent has a custom user defined, porterci, which has been configured with access to the Docker engine.
When the Azure Pipelines agent executes jobs, the jobs run under the porterci user account.

The vm is configured with environment variables and scripts so that Azure Pipelines can manage the virtual machine and start jobs.

## Maintaining the image
Right now we only update the image when it stops working for us.
For example, if we need a newer version of Docker Desktop installed, or need to adjust a configuration setting.
We do not regularly re-image the agent with security updates, and instead have the agent configured to install updates as needed.

NOTE: Only Microsoft employees can update the image, because our custom Windows agent infrastructure is all hosted in Azure on an internal subscription.

1. Log into the Azure subscription and locate the disk used to generate the current agent image.
   For example, porter-windows-agent-20220810.
2. Create a virtual machine from the disk using Standard D4s v3 (4 vcpus, 16 GiB memory).
   * It doesn't matter what you use for the admin account, since it will be removed when the vm is generalized later. 
     Use your name and preferred password.
   * Select the Windows Client model.
3. Log into the machine as the administrator account that you specified when you created the virtual machine.
   Use Bastion from inside your web browser to connect, not RDP.
4. To get into the Docker Desktop user interface, go to "Services" and first stop the "Docker Desktop" service.
   Then double-click on the Docker Desktop icon on the desktop to start a new instance with the user interface attached.
5. Make any necessary changes to the virtual machine.
6. Restart the machine and log in as porterci, validate that you can still run `docker ps`.
7. Shut down the machine and go to the virtual machine's disk in the Azure Portal.
   Create a snapshot of the disk named after the virtual machine.
   This snapshot is what you will use to create a vm the next time you need to update the agent image.
8. Start the machine and log in as the administrator and run the following command.
   ⚠️ The machine will log you out after the command runs, and you cannot log into the machine again afterwards!
   ```
   C:\Windows\System32\Sysprep\sysprep.exe /unattend:C:\unattend.xml /oobe /generalize /mode:vm /shutdown
   ```
9. Stop the vm, where the name will look like `porter-windows-agent-DATE` and DATE is YYYYMMDD.
   ```
   az vm stop --resource-group $RESOURCE_GROUP --name $NAME
   ```
10. Generalize the vm so that it can be used as a template for making new agents.
    ```
    az vm generalize --resource-group $RESOURCE_GROUP --name $NAME
    ```
11. Create a managed image from the vm.
    ```
    az image create --resource-group $RESOURCE_GROUP \
      --name $NAME --os-type windows \
      --source "/subscriptions/$SUBSCRIPTION/resourceGroups/$RESOURCE_GROUP/providers/Microsoft.Compute/virtualMachines/$NAME"
    ```
12. Update the virtual machine scale set to use the new managed image.
    ```
    az vmss update --resource-group $RESOURCE_GROUP \
      --name porter-windows \
      --set virtualMachineProfile.storageProfile.imageReference.id=/subscriptions/$SUBSCRIPTION/resourceGroups/$RESOURCE_GROUP/providers/Microsoft.Compute/images/$NAME
    ```
13. Update any existing agents to use the new image.
    ```
    az vmss update-instances --resource-group $RESOURCE_GROUP \
      --name porter-windows --instance-ids="*"
    ```
    This command takes about 15 minutes to complete.
    You can watch the progress by viewing the instances of the vmss in the portal.

## Initial Creation

These are only **notes** from when I initially created the first vm and vmss.
I don't remember all the steps anymore, but they may be helpful if we ever need to start over again.

**Create the virtual machine scale set**
```
az vmss create \
  --resource-group $RESOURCE_GROUP --name porter-windows \
  --image "/subscriptions/$SUBSCRIPTION/resourceGroups/$RESOURCE_GROUP/providers/Microsoft.Compute/images/$NAME" \
  --vm-sku Standard_D4s_v3 \
  --public-ip-per-vm \
  --admin-username porter --admin-password "$PASSWORD" \
  --instance-count 0 --disable-overprovision \
  --single-placement-group false --platform-fault-domain-count 1 \
  --upgrade-policy-mode manual --load-balancer ""
```

I think I could do this better by using an image gallery, then when the gallery is updated with a new image, the vmss would automatically use it.
But I had trouble getting that to work.

**Configure Azure Pipelines Agent**
See https://docs.microsoft.com/en-us/azure/devops/pipelines/agents/v2-windows?view=azure-devops

Allow powershell to run on the machine
```powershell
Set-ExecutionPolicy -ExecutionPolicy Unrestricted -Scope LocalMachine -Force
```

Change who the Azure Pipelines agent service runs as with unattended configuration (this is the user the jobs run as)
https://docs.microsoft.com/en-us/azure/devops/pipelines/agents/v2-windows?view=azure-devops#windows-only-startup
```powershell
[Environment]::SetEnvironmentVariable("VSTS_AGENT_INPUT_WINDOWSLOGONACCOUNT", "porterci", 'Machine')
[Environment]::SetEnvironmentVariable("VSTS_AGENT_INPUT_WINDOWSLOGONPASSWORD", "$PASSWORD", 'Machine')
```

I think this was a useful snippet for getting a service to run as a particular user but isn't the actual command that I ran
```
$svc=Get-CimInstance win32_service -Filter 'Name="browser"'
$svc|Invoke-CimMethod -MethodName Change -Arguments @{StartName='domain\user';StartPassword='Pass@W0rd'}
```

Configure the agent with the porter administrator account.
```
.\config.cmd --unattended --url https://dev.azure.com/getporter `
  --auth PAT --token $TOKEN --pool windows --agent manual-agent `
  --runasservice --windowslogonaccount porter --windowslogonpassword "$PASSWORD"
```
