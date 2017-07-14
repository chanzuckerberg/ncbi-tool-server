* NCBI-replica with version history data storage platform. Phase 1 of creating a service for accessing old versions of NCBI data. For use in the Infectious Disease platform.

* Components:
  * Sync service: https://github.com/chanzuckerberg/ncbi-tool-sync
  * Server service: https://github.com/chanzuckerberg/ncbi-tool-server
  * Command line client

* Planning docs:
  * Part 1: https://docs.google.com/document/d/1y9Y6Q5HgPHT5CfIPCMtkK2gIINtzcTEhdNzEWwqIIw4/edit
  * Part 2 and API documentation: https://docs.google.com/document/d/1mRzOFqJvhAWb4954o1eV-DVSvm_RFukohnt5bvTch-4/edit

* Testing:
  - To avoid running some of the acceptance tests, run go test with -short, e.g.
    - ```go test -short ./...```

- Folder structure for server component:
  - models/
    - directory.go
    - file.go
  - controllers/
    - application_controller.go
    - directory_controller.go
    - file_controller.go
  - utils/
    - context.go
    - utils.go
  - server.go
