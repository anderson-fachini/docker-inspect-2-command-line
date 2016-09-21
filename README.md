# docker-inspect-2-command-line
A nodejs script that generates the command line to start a docker image using the docker inspect of an existing image

## Usage
Save the docker inspect of you image in the root as a .json file and then run
```javascript
node index.js
```
You can save as many files as you want. The script will log each command line preceded by the name of the docker inspect file.
