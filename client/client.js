var imsApiClient = (function() {
  return {
    Config: function(apiAddr, authToken)
    {
      return {
        apiAddr: apiAddr,
        authToken: authToken,
      }
    },  
      
    // status : ongoing, published
    Project: function(name, projectName, description, status)
    {
      return {
        name: name,
        projectName: projectName,
        description: description,
        status: status,
      }
    },

    // status : draft, released
    // version: an integer
    Release: function(chapter, version, status)
    {
      return {
        chapter: chapter,
        version: version,
        status: status,
      }
    },

    // data : base64 encoded jpeg or png image
    Page: function(name, data)
    {
      return {
        page: name,
        data: data,
      }
    },

    getProjects: function(config, callback)
    {
      var elems = [config.apiAddr, 'projects']
      this.sendXMLHttpRequest(config, 'GET', elems.join('/'), '', callback)
    },

    createProject: function(config, data, callback)
    {
      var elems = [config.apiAddr, 'projects']
      this.sendXMLHttpRequest(config, 'POST', elems.join('/'), JSON.stringify(data), callback)
    },

    getProject: function(config, projectId, callback)
    {
      var elems = [config.apiAddr, 'projects', projectId]
      this.sendXMLHttpRequest(config, 'GET', elems.join('/'), '', callback)
    },

    updateProject: function(config, projectId, data, callback)
    {
      var elems = [config.apiAddr, 'projects', projectId]
      this.sendXMLHttpRequest(config, 'PUT', elems.join('/'), JSON.stringify(data), callback)
    },

    deleteProject: function(config, projectId, callback)
    {
      var elems = [config.apiAddr, 'projects', projectId]
      this.sendXMLHttpRequest(config, 'DELETE', elems.join('/'), '', callback)
    },

    createRelease: function(config, projectId, data, callback)
    {
      var elems = [config.apiAddr, 'projects', projectId, 'releases']
      this.sendXMLHttpRequest(config, 'POST', elems.join('/'), JSON.stringify(data), callback)
    },

    getReleases: function(config, projectId, callback)
    {
      var elems = [config.apiAddr, 'projects', projectId, 'releases']
      this.sendXMLHttpRequest(config, 'GET', elems.join('/'), '', callback)
    },

    getRelease: function(config, projectId, releaseId, callback)
    {
        var elems = [config.apiAddr, 'projects', projectId, 'releases', releaseId]
      this.sendXMLHttpRequest(config, 'GET', elems.join('/'), '', callback)
    },

    updateRelease: function(config, projectId, releaseId, data, callback)
    {
      var elems = [config.apiAddr, 'projects', projectId, 'releases', releaseId]
      this.sendXMLHttpRequest(config, 'PUT', elems.join('/'), JSON.stringify(data), callback)
    },

    deleteRelease: function(config, projectId, releaseId, callback)
    {
      var elems = [config.apiAddr, 'projects', projectId, 'releases', releaseId]
      this.sendXMLHttpRequest(config, 'DELETE', elems.join('/'), '', callback)
    },

    createPage: function(config, projectId, releaseId, data, callback)
    {
      var elems = [config.apiAddr, 'projects', projectId, 'releases', releaseId, 'pages']
      this.sendXMLHttpRequest(config, 'POST', elems.join('/'), JSON.stringify(data), callback)
    },

    getPages: function(config, projectId, releaseId, callback)
    {
      var elems = [config.apiAddr, 'projects', projectId, 'releases', releaseId, 'pages']
      this.sendXMLHttpRequest(config, 'GET', elems.join('/'), '', callback)
    },

    deletePage: function(config, projectId, releaseId, pageId, callback)
    {
      var elems = [config.apiAddr, 'projects', projectId, 'releases', releaseId, 'pages', pageId]
      this.sendXMLHttpRequest(config, 'DELETE', elems.join('/'), '', callback)
    },

    sendXMLHttpRequest: function(config, method, route, data, callback) {  
        if (typeof callback != 'function') {
          callback = function(r) {
            console.log(r)
          }
        }
        
        var xhr = new XMLHttpRequest()
        
        // asyhnchronous request
        xhr.open(method, route, true)
        
        xhr.onreadystatechange = function() {
          if(xhr.readyState === XMLHttpRequest.DONE) {
            if (xhr.status === 200) {
              callback(JSON.parse(xhr.responseText))
            } else {
              callback({ error: "unexpected status code", data: xhr })
            }
          }
        }
        
        //xhr.setRequestHeader('Auth-Token', config.authToken)

        // send the data
        xhr.send(data)
    },
  } // return
})() 