## archive-experiment standalone

The `archive-experiment` creates and/or restores one or more user defined archives. An archive is defined by a user specified collection of files to compress.  An archive can be restored by specifying the archive to restore along with with the output location of restored files.  Since archiving is dependent on file size and the number of files, it can take a long time which would be unsatisfactory to a Web UI user.  Consequently, archiving an experiment when an experiment is stopped could lock up the UI with no feedback mechanism except for errors.  It would be nice to extend the Phenix user app lifecycle to allow for status updates and run in a separate thread to accomodate long running user apps.  The restore feature of this user app is meant to restore an experiment's initial state from an archive.  Restoring an experiment is the same as creating an experiment.  Consequently, restoration of a previous experiment did not fit well in the Phenix user app lifecycle.

Since this user app does not fit well with the current Phenix user app lifecyle, a standalone version of this app was created.  It only receives two arguments namely the `experiment name` and the `lifecycle stage`.  The `cleanup` lifecycle is used for archiving while `configure` lifestyle is used for restoring an experiment.  An example invocation is shown below:

`phenix-app-archive-experiment --experiment test --stage cleanup`

The use app does need to know the location of the Phenix datastore.  Consequently, define and pass the `PHENIX_STORE_ENDPOINT` environment variable or the default location of `bolt://etc/phenix/store.bdb` will be used.  

In addition, an environment variable for the log file should be defined as this user app makes use of a log file.  The default location is `/var/log/phenix/phenix.log` if the `PHENIX_LOG_FILE` environment variable is not defined.

### archive

An archive is defined by specifying the following five fields under the `archives` list:    

**name**: The archive name will have either a `tar.gz` or a `zip` extension appended based on the archive type specified for `type`.  The archive name can also consist of two place holder variables namely `<experiment_name>` and `<timestamp>`.  The `<experiment_name>` placeholder will receive the name of the experiment and the `<timestamp>` placeholder will receive a timestamp in the format `YYYY-MM-DD_HHMM`.  For example if a user specified the name `<experiment_name>_test_<timestamp>` and the name of the experiment was `archive` and the current date and time was `2021-03-10 16:10`, then the name of the archive would be `archive_test_2021-03-10_1610`.  The default value is `<experiment_name>_<timestamp>` when no value is specified.  
**directory**: Directory is the path to the parent directory that should be scanned to collect files to archive.  The name `experiment_directory` can be used to refer to the directory where experiment files are usually placed. `(e.g. /phenix/images/{experiment name}/files)`  
**filters**: Filters are a list of regular expressions so that only files matching the regular expressions will be included in the archive.  This field is optional.  If no filters are specified, then all files found in the specified `directory` will be included in the archive.  
**cleanup**: `[true,false]` Cleanup specifies whether the files collected in the archive should be removed after successfully placed in an archive.  A value of `true` will remove the files from the `directory` specified.  The default value is `false`.  
**type**: [targz,zip]  Type describes the algorithms used to construct the archive.  If type is `targz`, then the archive will be a `tar` using `gzip` compression.  If `zip` is specified, then the archive will be a zip archive.  The default value is `zip`.  
**output**: Output describes the directory where the archive should be written to.  The default value is `/phenix/Archives` when no output directory is specified.  The placeholder variables `<experiment_name>` and `<timestamp>` can also be used as part of the output directory name.  If the output directory does not exist, it will be created.  

Below are several examples of creating an archive.  

### Minimum archive

If only the `archive-experiment` app is included with an empty `archives` list, then the experiment configuration files for topology, scenario, and experiment will be archived with the default archive name in the default archive output location.  In fact, an archive with the experiment configuration files will always be created when the `archive-experiment` app is included with an `archives` list.  

```
apiVersion: phenix.sandia.gov/v1
kind: Scenario
metadata:
  name: archive_test_empty
  annotations:
    topology: inf_topo_no_mgnt
spec:
  apps:
    experiment:        
    - name: archive-experiment
      metadata:
        archives:

```

### Archive with placeholder variables

```
apiVersion: phenix.sandia.gov/v1
kind: Scenario
metadata:
  name: archive_variable_test
  annotations:
    topology: inf_topo_no_mgnt
spec:
  apps:
    experiment:        
    - name: archive-experiment
      metadata:
        archives:
          - name: <experiment_name>_test_<timestamp>
            directory: experiment_directory
            filters:
            cleanup:
            type:
            output:
          
    host:

```

### Archive with a simple filter

```
apiVersion: phenix.sandia.gov/v1
kind: Scenario
metadata:
  name: archive_filter_test
  annotations:
    topology: inf_topo_no_mgnt
spec:
  apps:
    experiment:        
    - name: archive-experiment
      metadata:
        archives:
          - name: <experiment_name>_test_<timestamp>
            directory: experiment_directory
            filters: 
              - pcap
            cleanup: true
            type: zip
            output: /phenix/Archives/filter_test_<timestamp>
          
    host:

```

### Multiple archives

```
apiVersion: phenix.sandia.gov/v1
kind: Scenario
metadata:
  name: archive_multiple_test
  annotations:
    topology: inf_topo_no_mgnt
spec:
  apps:
    experiment:        
    - name: archive-experiment
      metadata:
        archives:
          - name: <experiment_name>_test_<timestamp>
            directory: experiment_directory
            filters: 
              - pcap
            cleanup: true
            type: zip
            output: /phenix/Archives/experiment_files_<timestamp>            
          - name: data_<timestamp>
            directory: /phenix/data
            filters: 
              - cfg
            cleanup: true
            type: targz
            output: 
          
    host:

```

### Restore an archive

Restoration of an archive consists of creating an experiment using a saved experiment configuration file and restoring any user defined files.  The restored experiment will have the name `<experiment_name>_<saved_date_time>`.  If the `<saved_date_time>` can not be extracted from the archive name, then the original experiment name will be used.  If the original experiment name already exists in the experiment data store, then the experiment will not be restored.  Perhaps, in the future, an option to delete an existing experiment will be added to the restoration specification.  The restoration of an experiment will occur during the `configure` stage as that was the best option with regards to the Phenix user app lifecycle.

An archive can be restored by specifying the following three fields under the `retrievals` list:  

**name**: This should specify the full path to the archive.  
**directory**: Directory is the output location for the restored files.  If no directory is specified, then the files will be restored to directories stored in the archive.  If configuration files are specified in the archive with no directory references, then the default location is `/phenix/configurations`  
**filters**: Filters are a list of regular expressions so that only files matching the regular expressions will be restored from the archive.  This field is optional.  If no filters are specified, then all files found in the specified archive will be restored.  


```
apiVersion: phenix.sandia.gov/v1
kind: Scenario
metadata:
  name: archive_restore_test
  annotations:
    topology: inf_topo_no_mgnt
spec:
  apps:
    experiment:        
    - name: archive-experiment
      metadata:
        retrievals:
          - name: /phenix/Archives/archive_filters_test_2021-03-10_1510/archive_filters_test_2021-03-10_1510.tar.gz
            directory: 
            filters:  
          
    host:

```



   
