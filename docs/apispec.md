# IMangaScans API

## Convention

Unless specified otherwise, all responses will be served as JSON.  The common exception to this rule will be in cases where files are served rather than informational responses.

`GET` and `DELETE` requests will accept parameters in URL query string parameters.

`POST` and other types of requests will accept parameters as JSON in the body of the request.

JSON responses will always contain an `error` field.  If an error occurred in processing the request, it will contain a string describing the error. Otherwise, `error` will be `null`.

## Notation

Types will be represented with the name of a type, such as `string`, `bool`, `integer`, `float` etc.

Compound types will be represented as `[<type>]` for arrays with values of all the same type, `[<type1>, <type2>, ... <typeN>]` in the case of an array with exactly N values whose types are predefined, and `{"key1": <type>, "key2": <type>}` in the case of objects. Objects may also have the type `{...}`, indicating that they can contain any combination of keys and values.

Any parameter that is optional will have a type prefixed with `optional`. For example, the type `optional integer` indicates that the field is not required but, if provided, must be an integer.

## Endpoints

### Download an archive of a release

```
GET /{projectName}-{chapter}{groupName}{checksum}.{version}.zip
```

* Downloads the release archive
* The route is invalid if the release is not in released state
* The archive may not contain any folders
* The archive must contain a credit page.


#### Parameters

Instead of this endpoint accepting parameters in the usual URL query parameter format, the name of the zip file requested contains positional arguments in the URL.

Name | Type | Description
-----|------|------------
projectName | string | The name of a project
chapter | integer | The id of the release
groupName | string | The shorthand name of the scanlation group, e.g. ims
checksum | string | The hex encoding of the sha256 hash of the zip file
version | integer | The version number of the release, starting from 1

#### Response

* Status 200: The zip file will be served directly
* Status 4xx: Invalid request, with a string error message
* Status 5xx: Server error, with a string error message

### Download a single image

```
GET /{projectName}-{chapter}.{version}/{page}.{ext}
```

* Downloads the requested image from the archive
* The route is valid even if the release is not in released state

#### Parameters

Instead of this endpoint accepting parameters in the usual URL query parameter format, the name of the image file requested contains positional arguments in the URL.

Name | Type | Description
-----|------|------------
projectName | string | The name of a project
chapter | integer | The chapter number
version | integer | The version number of the release, starting from 1
page | integer | The page number, starting from 1
ext | string | The filetype extension of the page's file

#### Response

* Status 200: The image file will be served directly
* Status 4xx: Invalid request, with a string error message
* Status 5xx: Server error, with a string error message

### Get a list of all projects

```
GET /projects
```

#### Parameters

Name | Type | Description
-----|------|------------
ordering | optional string | Either "newest" or "oldest" for newest first or oldest first respectively

#### Response

Name | Type | Description
-----|------|------------
projects | `[{"id": string, "name": string, "createdAt": string}]` | An array containing objects that identify all existing projects

### Create a new project

```
POST /projects
```

#### Parameters

Name | Type | Description
-----|------|------------
name | string | The human-readable name of the project
filename | string | The projectName identifier for archive filenames. Must be unique
description | string | A longer description of the project
status | string | The current status of the project. One of "ongoing", "complete", "dropped", etc.

#### Response

Name | Type | Description
-----|------|------------
success | bool | True if there was no problem creating the project
id | integer | The id of the newly created project, if successful

### Get information about a project

```
GET /projects/{projectId}
```

#### Parameters

Name | Type | Description
-----|------|------------
projectId | integer | The identifier of a project, as returned by the create endpoint

#### Response

Name | Type | Description
-----|------|------------
createdAt | string | The date when the project was created
name | string | The descriptive name of the project
filename | string | The projectName
status | string | The current status of the project
description | string | The description of the project

### Update project information

```
PUT /projects/{projectId}
```

#### Paramters

Name | Type | Description
-----|------|------------
projectId | integer | The unique identifier for the project
name | optional string | A new human-readable name for the project
filename | optional string | A new projectName for the project
status | optional string | The new status of the project
description | optional string | A new description for the project

#### Response

Name | Type | Description
-----|------|------------
success | bool | True if the update could take place, else false

### Delete a project

```
DELETE /projects/{projectId}
```

Also removes all of the releases associated with the project

#### Parameters

Name | Type | Description
-----|------|------------
projectId | integer | The unique identifier for the project

#### Response

Name | Type | Description
-----|------|------------
success | bool | True if the project could be deleted, else false

### Get a list of releases for a project

```
GET /projects/{projectId}/releases
```

By default, will be ordered by newest to oldest chapter.

#### Parameters

Name | Type | Description
-----|------|------------
projectId | integer | The unique identifier for the project
ordering | optional string | Either "newest" or "oldest" to specify whether the newest or oldest releases should be listed first

#### Response

Name | Type | Description
-----|------|------------
releases | `[{"id": integer, "releasedOn": string, "version": integer, "checksum": string}]` | An array of objects containing information about releases

### Create a new release

```
POST /projects/{projectId}/releases
```

#### Parameters

Name | Type | Description
-----|------|------------
chapter | integer | The number of the chapter
version | integer | The version of the release of the particular chapter
status | string | The release status, e.g. "released" or "draft"

#### Response

Name | Type | Description
-----|------|------------
success | bool | True if the release could be made, else false
id | integer | The newly created id of the release if successful

### Get information about a release

```
GET /projects/{projectId}/releases/{releaseId}
```

#### Parameters

Name | Type | Description
-----|------|------------
projectId | integer | The unique id of the project under which the release was created
releaseId | integer | The unique id of the release

#### Response

Name | Type | Description
-----|------|------------
projectName | string | The unique projectName of the project that the release was created for
chapter | integer | The integer number of the chapter the release contains
groupName | string | "ims"
checksum | string | The hex encoding of the sha256 checksum of the release's zip file
version | integer | The release version number
status | string | The status of the release

### Update information about a release

```
PUT /projects/{projectId}/releases/{releaseId}
```

#### Parameters

Name | Type | Description
-----|------|------------
projectId | integer | The unique id of the project under which the release was created
releaseId | integer | The unique id of the release
chapter | optional integer | The new chapter number
version | optional integer | The new version number for the release
status | optional string | The new status of the release

#### Response

Name | Type | Description
-----|------|------------
success | bool | True if the update could take place successfully, else false

### Delete a release

```
DELETE /projects/{projectId}/releases/{releaseId}
```

Deletes a release and all associated pages.

#### Parameters

Name | Type | Description
-----|------|------------
projectId | integer | The unique id of the project under which the release was created
releaseId | integer | The unique id of the release

#### Response

Name | Type | Description
-----|------|------------
success | bool | True if the project could be deleted, else false

### Get a list of pages in a release

```
GET /projects/{projectId}/releases/{releaseId}/pages
```

#### Parameters

Name | Type | Description
-----|------|------------
projectId | integer | The unique id of the project under which the release was created
releaseId | integer | The unique id of the release

#### Response

Name | Type | Description
-----|------|------------
pages | `[{"projectName": string, "chapter": integer, "version": integer, "page": integer, "ext": string}]`| An array of objects describing each of the pages part of the release

### Add a new page to a release

```
POST /projects/{projectId}/releases/{releaseId}/pages
```

#### Parameters

Name | Type | Description
-----|------|------------
projectId | integer | The unique id of the project under which the release was created
releaseId | integer | The unique id of the release
page | integer | The page number

#### Release

Name | Type | Description
-----|------|------------
success | bool | True if the page could be uploaded, else false e.g. if the release has a completed status
id | integer | The unique id of the newly created page

### Delete a page from a release

```
DELETE /projects/{projectId}/releases/{releaseId}/pages/{pageId}
```

#### Parameters

Name | Type | Description
-----|------|------------
projectId | integer | The unique id of the project under which the release was created
releaseId | integer | The unique id of the release
pageId | integer | The unique identifier of the page to delete

#### Response

Name | Type | Description
-----|------|------------
success | bool | True if the page could be deleted, else false e.g. if the release has already been moved to a complete status