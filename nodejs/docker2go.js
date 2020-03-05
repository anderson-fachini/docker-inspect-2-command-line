const fs = require('fs')
const path = require('path')
const argv = require('yargs').argv

const args = argv._

let format = argv.format

function jsonInspectToCommand(fileContent) {
    const dockerConfig = JSON.parse(fileContent)[0]
    const fmtEnding = format ? ' \\\n' : ''

    let command = ''
    command += 'docker run -d'
    command += ' --name=' + dockerConfig.Name.substr(1)
    command += fmtEnding

    if (dockerConfig.HostConfig.AutoRemove) {
      command += " --rm";
    }

    if (dockerConfig.HostConfig.RestartPolicy.Name != 'no' && dockerConfig.HostConfig.RestartPolicy.Name != '') {
      command += ' --restart ' + dockerConfig.HostConfig.RestartPolicy.Name
      command += fmtEnding
    }

    let isHexa = /^[0-9a-fA-F]+$/.test(dockerConfig.Config.Hostname)
	  if (!isHexa) {
      command += " --hostname=" + dockerConfig.Config.Hostname
      command += fmtEnding
    }

    if (dockerConfig.HostConfig.Memory > 0) {
      // memory is stored in bytes, minumum is 4M
      command += ' -m '

      var kbMemory = parseInt(dockerConfig.HostConfig.Memory, 10) / 1024
      var mbMemory = kbMemory / 1024
      var gbMemory = mbMemory / 1024

      if (gbMemory >= 1 && Math.round(gbMemory) == gbMemory) {
        command += Math.round(gbMemory) + 'g'
          }
          else if (mbMemory >= 1 && Math.round(mbMemory) == mbMemory) {
        command += Math.round(mbMemory) + 'm'
          }
          else if (kbMemory >= 1 && Math.round(kbMemory) == kbMemory) {
        command += Math.round(kbMemory) + 'k'
          }
          else {
        command += dockerConfig.HostConfig.Memory + 'b'
      }
      command += fmtEnding
    }

    if (dockerConfig.HostConfig.NanoCpus > 0) {
      var cpus = dockerConfig.HostConfig.NanoCpus / 1000000000
		  command += " --cpus=" + cpus
    }

    if (dockerConfig.HostConfig.Dns != null) {
      for (let i=0; i < dockerConfig.HostConfig.Dns.length; i++) {
        command += ' --dns=' + dockerConfig.HostConfig.Dns[i]
        command += fmtEnding
      }
    }

    if (dockerConfig.Config.Env != null) {
      for (let i=0; i < dockerConfig.Config.Env.length; i++) {
              let env = dockerConfig.Config.Env[i]

        if (!(env.startsWith('PATH') || env.startsWith('LANG') || env.startsWith('LC_ALL'))) {
                  command += ' -e '

          if (env.indexOf(' ') != -1) {
            command += env.substr(0, env.indexOf('=')) + '="' + env.substr(env.indexOf('=')+1) + '"'
          } else {
            command += env
          }
          command += fmtEnding
        }
      }
    }

    if (dockerConfig.HostConfig.Binds != null) {
      for (let i=0; i < dockerConfig.HostConfig.Binds.length; i++) {
              let bind = dockerConfig.HostConfig.Binds[i]

        command += ' -v '
        if (bind.indexOf(' ') != -1) {
          bind = '"' + bind + '"'
        }
        command += bind
        command += fmtEnding
      }
    }

    if (dockerConfig.HostConfig.PortBindings != null) {
      for (let port in dockerConfig.HostConfig.PortBindings) {
              let binding = dockerConfig.HostConfig.PortBindings[port]

        for (let i=0; i < binding.length; i++) {
          command += ' -p ' + binding[i].HostPort + ':' + port.split('/')[0]
          command += fmtEnding
        }
      }
    }

    if (dockerConfig.HostConfig.Links != null) {
      for (let i=0; i < dockerConfig.HostConfig.Links.length; i++) {
              let splitted = dockerConfig.HostConfig.Links[i].split('/')

        let preLink = splitted[1].substr(0, splitted[1].length-1)
              command += ' --link ' + preLink

        if (preLink != splitted[3]) {
          command += ':' + splitted[3]
        }

        command += fmtEnding
      }
    }

    if (Object.keys(dockerConfig.HostConfig.LogConfig.Config).length > 0) {
      for (let config in dockerConfig.HostConfig.LogConfig.Config) {
        command += ' --log-opt ' + config + '=' + dockerConfig.HostConfig.LogConfig.Config[config]
      }
      command += fmtEnding
    }

    command += ' ' + dockerConfig.Config.Image

    return command
}

if (args.length > 0) {
    for (let i = 0; i < args.length; i++) {
        const fileName = path.resolve(args[i])

        if (!fs.statSync(fileName).isDirectory()) {
            const fileContent = fs.readFileSync(fileName, 'utf8')

            const command = jsonInspectToCommand(fileContent)

            if (args.length > 1) {
                console.log("File:", path.basename(fileName))
                console.log(command)

                if (i < (args.length - 1)) {
                    console.log()
                }
            } else {
                console.log(command)
            }
        }
    }
}
else if (!process.stdin.isTTY) {
    process.stdin.on( 'data', function(data) {
        let command = jsonInspectToCommand(data)
        console.log(command)
    });
}
else {
    console.log('Usage:')
    console.log('./docker2go docker_inspect_of_a_container.txt OR')
    console.log('docker inspect <container> | docker2go')
}