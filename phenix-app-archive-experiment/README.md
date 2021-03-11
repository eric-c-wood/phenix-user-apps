## archive-experiment

The `archive-experiment` creates and/or restores one or more user defined archives. An archive is defined by a user specified collection of files to compress.  An archive can be restored by specifying the archive to restore along with with the output location of restored files.  An archive will be created when an experiment is stopped from the Web UI.  Archives will be restored when an experiment is created.  Ideally, the restoration of an archive will create an experiment rather than running when an experiment is being created.  The `restore` feature of this app does not fit well into the current phenix user app lifecycle and may be best implemented as a standalone executable.  

### archive

An archive is defined by specifying the following five fields under the `archives` list:    

**name**: The archive name will have either a `tar.gz` or a `zip` extension appended based on the archive type specified for `type`.  The archive name can also consist of two place holder variables namely `<experiment_name>` and `<timestamp>`.  The `<experiment_name>` placeholder will receive the name of the experiment and the `<timestamp>` placeholder will receive a timestamp in the format `YYYY-MM-DD_HHMM`.  For example if a user specified the name `<experiment_name>_test_<timestamp>` and the name of the experiment was `archive` and the current date and time was `2021-03-10 16:10`, then the name of the archive would be `archive_test_2021-03-10_1610`.  The default value is `<experiment_name>_<timestamp>` when no value is specified.  
**directory**: Directory is the path to the parent directory that should be scanned to collect files to archive.  The name `experiemnt_directory` can be used to refer to the directory where experiemnt files are usually placed. `(e.g. /phenix/images/{experiment name}/files)`  
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

Restoration of an archive consists of creating an experiement using a saved experiment configuration file and restoring any user defined files.  The restored experiment will have the name `<experiment_name>_<saved_date_time>`.  If the `<saved_date_time>` can not be extracted from the archive name, then the original experiment name will be used.  If the original experiment name already exists in the experiment data store, then the experiment will not be restored.  Perhaps, in the future, an option to delete an existing experiment will be added to the restoration specification.  As mentioned, the restoration option does not fit well with the current phenix experiment lifecycle.  Consequently, the restoration feature of this app will most likely be moved to a standalone executable.  

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



   
