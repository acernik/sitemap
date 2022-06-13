## Sitemap
This repository contains implementation of a simple sitemap generator (`https://www.sitemaps.org`).

## Functionality
The sitemap generator accepts the following parameters:
- `-url` - the starting URL for the sitemap generator
- `-parallel` - number of parallel workers to navigate through site
- `-output-file` - output file for the sitemap
- `-max-depth` - max depth of url navigation recursion

To run the application run `make run` command. This will run the sitemap generator with parameters defined in `run.sh` 
file. If you want to change the parameters, then make the changes in `run.sh` file. Or just run `go run cmd/main.go` from 
root directory with parameters of your choice.

## Unit tests
To run unit tests execute `make tests` command from the root directory. This will run all unit tests with coverage.