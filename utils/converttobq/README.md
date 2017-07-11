# Convert data into BigQuery ingest-able format using a Generic Data Converter

## Running Locally
* Create a BigQuery dataset `reposupervisords` and an empty table `reposupervisor_<target>` in Google BigQuery from the GCP UI to store the processed repo-supervisor results with the following schema (all nullable):
```
File:string
Secret:string
```

* Similarly for wfuzz, create the dataset `wfuzzds` and an empty table `wfuzz_<appServerTech>_<target>` in Google BigQuery from the GCP UI to store the processed wfuzz results with the following schema (all nullable):
```
ID:string
Response:string
Lines:string
Word:string
Chars:string
Request:string
Success:string
```

* Run Repo Supervisor by typing `docker run -it abhartiya/tools_gitallsecrets:v3 -token <> -org <> -toolName repo-supervisor`. Copy the `/data/results.txt` from the container to `results.json`.

* Run WFUZZ by typing `docker run -it abhartiya/tools_wfuzz -w /data/SecLists/Discovery/Web_Content/tomcat.txt --hc 404,429,400 -o csv <URL>/FUZZ /data/out.csv`. Copy the `/data/out.csv` from the container to `out.csv`.

* Complete the `.env.sample` file in the `data-converter` folder with the appropriate values and rename it to `.env`.

* Now, in order to convert Repo Supervisor's output and upload it to BQ, type `go run dataconvert.go -toolName repo-supervisor -filePath results.json`

* Next, in order to convert WFUZZ's output and upload it to BQ, type `go run dataconvert.go -toolName wfuzz -filePath out.csv`