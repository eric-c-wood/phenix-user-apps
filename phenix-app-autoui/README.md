## autoui

Phenix user app that generates a YAML configuration file from a Phenix topology and associated YAML/Golang template files.  The autoui user app will automatically launch the auto-ui application with the generated configuration file to faciliate executing a list commands/actions for one or more virtual machines.  On experiment `cleanup`, all running instances of auto-ui will be stopped.   

### configuration

To apply templates to a virtual machine, add `ui-scripts` under `labels` in the topology.  The values for the `ui-scripts` key will be a comma separated list of template names without the extension.  The templates are Golang and leverage the Phenix schema.





