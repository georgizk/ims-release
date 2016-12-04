// client and jquery need to be loaded
var imsApiApp = function() {
  var projectsTableId = '#projects'  
  var releasesTableId = '#releases'
  var pagesTableId = '#pages'
  
  var projectEditFormId = '#projectEditForm'
  var projectDeleteId = '#projectDeleteConfirm'
  var projectAddFormId = '#projectAddForm'
  
  var releaseEditFormId = '#releaseEditForm'
  var releaseDeleteId = '#releaseDeleteConfirm'
  var releaseAddFormId = '#releaseAddForm'
  
  var pageDeleteId = '#pageDeleteConfirm'
  var pageAddFormId = '#pageAddForm'
  
  return {
    createCallback: function(fieldName, tableId) {
      return function(data) {
        if (data.error) {
          console.log('got error')
          console.log(data)
          return
        }    
        var rows = data[fieldName]
        var table = $(tableId).DataTable()
        table.clear()
        table.rows.add(rows)
        table.draw()
      }
    },
    
    createSelectionCallback: function(tableId, cbSelected) {
      $(tableId + ' tbody').on('click', 'tr', function() {
        if($(this).hasClass('selected')) {
          $(this).removeClass('selected')
        }
        else {
          $(tableId + ' tr.selected').removeClass('selected')
          $(this).addClass('selected')
          cbSelected($(this))
        }
      })
    },
    
    createSubmitCallback: function(cbSubmit) {
      return function(e) {
        e.preventDefault()
        data = $(this).serializeArray()
        mappedData = {}
        for (var field of data) {
          mappedData[field.name] = field.value
        }
        cbSubmit(mappedData)
      }                
    },
    
    initTables: function(config) {
      var projectsTable = $(projectsTableId).DataTable({
        data: [],
        columns: [
            { data: 'id' },
            { data: 'name' },
            { data: 'projectName' },
            { data: 'description' },
            { data: 'status' },
            { data: 'createdAt' },
        ]
      })
      
      var projectAddForm = $(projectAddFormId)
      projectAddForm.submit(imsApiApp.createSubmitCallback(function(data) {
        imsApiClient.createProject(config, data, function(r) {
          console.log(r)
          imsApiClient.getProjects(config, imsApiApp.createCallback('projects', projectsTableId))
          $.fancybox.close()
        })
      }))
      
      imsApiApp.createSelectionCallback(projectsTableId, function(row) {
        var cols = row.find('td')
        var id = $(cols[0]).text()
        console.log("loading releases for project " + id)
        var table = $(releasesTableId).DataTable()
        table.clear()
        imsApiClient.getReleases(config, id, imsApiApp.createCallback('releases', releasesTableId))
        
        var editForm = $(projectEditFormId)
        var fields = editForm.find(':input')
        $(fields[0]).val($(cols[1]).text())
        $(fields[1]).val($(cols[2]).text())
        $(fields[2]).val($(cols[3]).text())
        $(fields[3]).val($(cols[4]).text())
        
        editForm.off('submit')
        editForm.submit(imsApiApp.createSubmitCallback(function(data) {
          imsApiClient.updateProject(config, id, data, function(r) {
            console.log(r)
            imsApiClient.getProjects(config, imsApiApp.createCallback('projects', projectsTableId))
            $.fancybox.close()
            $(projectEditFormId).off('submit')
          })
        }))
        
        var addForm = $(releaseAddFormId)
        addForm.off('submit')
        addForm.submit(imsApiApp.createSubmitCallback(function(data) {
          data.version = parseInt(data.version)
          imsApiClient.createRelease(config, id, data, function(r) {
            console.log(r)
            imsApiClient.getReleases(config, id, imsApiApp.createCallback('releases', releasesTableId))
            $.fancybox.close()
            $(releaseAddFormId).off('submit')
          })
        }))    
        
        var deleteButton = $(projectDeleteId)
        deleteButton.off('click')
        deleteButton.click(function(e) {
          e.preventDefault()
          imsApiClient.deleteProject(config, id, function(r) {
            console.log(r)
            imsApiClient.getProjects(config, imsApiApp.createCallback('projects', projectsTableId))
            $.fancybox.close()
            $(projectDeleteId).off('click')
          })
        })
      })
      
      var releasesTable = $(releasesTableId).DataTable({
        data: [],
        columns: [
            { data: 'id' },
            { data: 'projectId' },
            { data: 'chapter' },
            { data: 'version' },
            { data: 'status' },
            { data: 'releasedOn' },
        ]
      })
      
      imsApiApp.createSelectionCallback(releasesTableId, function(row) {
        var cols = row.find('td')
        var id = $(cols[0]).text()
        var projectId = $(cols[1]).text()
        
        console.log("loading pages for project " + projectId + " and release " + id)
        var table = $(pagesTableId).DataTable()
        table.clear()
        imsApiClient.getPages(config, projectId, id, imsApiApp.createCallback('pages', pagesTableId))     

        var editForm = $(releaseEditFormId)
        var fields = editForm.find(':input')
        $(fields[0]).val($(cols[2]).text())
        $(fields[1]).val($(cols[3]).text())
        $(fields[2]).val($(cols[4]).text())
        
        editForm.off('submit')
        editForm.submit(imsApiApp.createSubmitCallback(function(data) {
          data.version = parseInt(data.version)
          console.log(data)
          imsApiClient.updateRelease(config, projectId, id, data, function(r) {
            console.log(r)
            imsApiClient.getReleases(config, projectId, imsApiApp.createCallback('releases', releasesTableId))
            $.fancybox.close()
            $(releaseEditFormId).off('submit')
          })
        }))    
        
        var addForm = $(pageAddFormId)
        addForm.off('submit')
        addForm.submit(imsApiApp.createSubmitCallback(function(data) {
          console.log(data)
          var fileInput = document.getElementById('pagesToUpload')
          var files = fileInput.files
          var fileIdx = 0
          var onloadFunction = function(fileIdx, files)
          {
            return function(e) {
              console.log('fired')
              console.log(fileIdx)
              console.log(files)
              var base64url = e.target.result
              var firstComma = base64url.indexOf(',')              
              var base64enc = base64url.substr(firstComma + 1)
              var page = imsApiClient.Page(files[fileIdx].name, base64enc)
              console.log(page)
              imsApiClient.createPage(config, projectId, id, page, function() {
                imsApiClient.getPages(config, projectId, id, imsApiApp.createCallback('pages', pagesTableId))
                fileIdx++
                if (fileIdx < files.length) {
                  e.target.onload = onloadFunction(fileIdx, files)
                  e.target.readAsDataURL(files[fileIdx])
                } else {                
                  $.fancybox.close()
                }
              })
            }
          }
          
          var fileReader = new FileReader()
          if (fileIdx < files.length) {
            fileReader.onload = onloadFunction(fileIdx, files)
            fileReader.readAsDataURL(files[fileIdx])
          }
        })) 

        var deleteButton = $(releaseDeleteId)
        deleteButton.off('click')
        deleteButton.click(function(e) {
          e.preventDefault()
          imsApiClient.deleteRelease(config, projectId, id, function(r) {
            console.log(r)
            imsApiClient.getReleases(config, projectId, imsApiApp.createCallback('releases', releasesTableId))
            $.fancybox.close()
            $(releaseDeleteId).off('click')
          })
        })
      })
      
      var pagesTable = $(pagesTableId).DataTable({
        data: [],
        columns: [
            { data: 'id' },
            { data: 'releaseId' },
            { data: 'page' },
            { data: 'createdAt' }
        ]
      })
      
      imsApiApp.createSelectionCallback(pagesTableId, function(row) {
        var deleteButton = $(pageDeleteId)
        deleteButton.off('click')
        
        var cols = row.find('td')
        var pageId = $(cols[0]).text()
        var slectedProjectCols = $(projectsTableId + ' tr.selected td')
        var projectId = $(slectedProjectCols[0]).text()
        var releaseId = $(cols[1]).text()
        deleteButton.click(function(e) {
          e.preventDefault()
          imsApiClient.deletePage(config, projectId, releaseId, pageId, function(r) {
            console.log(r)
            imsApiClient.getPages(config, projectId, releaseId, imsApiApp.createCallback('pages', pagesTableId))
            $.fancybox.close()
            $(releaseDeleteId).off('click')
          })
        })
      })
      
      imsApiClient.getProjects(config, imsApiApp.createCallback('projects', projectsTableId))
    },
  }
}()

$(document).ready(function() {
  var apiAddr = prompt('Please enter the api address')
  var config = imsApiClient.Config(apiAddr)
  
  $('a.overlay').click(function(e) {
    // only on left click
    if (e.which != 1) return;
    
    $.fancybox([{
      href: $(this).attr('href'),
      padding: 0, // this disables the ugly padding
      helpers: {
        overlay: {
          locked: false // this disables the scrolling to the top
        }
      }
    }])   
    // prevent hyperlink from activating
    return false; 
  })
  
  $('button.cancel').click(function(e) {
    $.fancybox.close()
  })
  imsApiApp.initTables(config)
});