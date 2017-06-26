* NCBI-replica with version history data storage platform. Phase 1 of creating a service for accessing old versions of NCBI data.

* See planning document: https://docs.google.com/document/d/1y9Y6Q5HgPHT5CfIPCMtkK2gIINtzcTEhdNzEWwqIIw4/edit

* Testing:
  - To avoid running some of the acceptance tests, run go test with -short, e.g.
    - ```go test -short ./...```

- Folder structure for server component:
  - config.yaml (Config file)
  - models/
  - controllers/
  - utils/
  - web/
