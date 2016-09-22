var fs = require('fs')

fs.readdir('.', function(err, files) {
    if (err) {
        console.log(err)
    } else {
        files.forEach(function(file) {
            if (file.endsWith('.json')) {
                console.log(file);
                console.log(getDockerCommandLine(file))
                console.log()
            }
        });
    }
});

/**
 * If a environment variable value has one of this characters, it needs to be wrapped with double quotation
 */
function hasWrapChar(str) {
	var wrapChars = [' ', '&', ';'];
	for (var c in wrapChars)
		if (str.indexOf(wrapChars[c]) !== -1)
			return true;
	return false;
}

function getDockerCommandLine(fileName) {
    var dockerConfig = require('./' + fileName)[0];
    var command = 'docker run -d';
    command += ' --name=' + dockerConfig['Name'].substring(1);

    function getEnvs(dockerConfig) {
        var ret = '';

        var envs = dockerConfig['Config']['Env'];
        if (envs != null) {
            for (var i = 0; i < envs.length; i++) {
                var env = envs[i];
                if (!env.startsWith('PATH') && !env.startsWith('LANG') && !env.startsWith('LC_ALL')) {
                    ret += ' -e ';
                    if (hasWrapChar(env)) {
                        ret += env.substring(0, env.indexOf('=')) + '="' + env.substring(env.indexOf('=')+1) + '"';
                    } else {
                        ret += env;
                    }
                }
            }
        }

        return ret;
    }

    function getHostConfig(dockerConfig) {
        var ret = '';

        if (Object.keys(dockerConfig['HostConfig']['PortBindings']).length > 0) {
            for (var port in dockerConfig['HostConfig']['PortBindings']) {
                var bindings = dockerConfig['HostConfig']['PortBindings'][port];
                for (var i = 0; i < bindings.length; i++) {
                    ret += ' -p ' + bindings[i]['HostPort'] + ':' + port.split('/')[0];
                }
            }
        }

        if (dockerConfig['HostConfig']['Links'] != null) {
            var links = dockerConfig['HostConfig']['Links'];
            for (var i = 0; i < links.length; i++) {
                var link = links[i];
                var splitted = link.split('/');
                ret += ' --link ' + splitted[1] + splitted[3];
            }
        }

        return ret;
    }


    command += getEnvs(dockerConfig);
    command += getHostConfig(dockerConfig);

    command += ' ' + dockerConfig['Config']['Image'];

    return command;
}
