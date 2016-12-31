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

### Get a list of all projects

```
GET /projects
```

#### Parameters

None

#### Response

Name | Type | Description
-----|------|------------
projects | `[{"id": string, "name": string, "shorthand": string, "description": string, "status": string, "createdAt": string}]` | An array containing objects that identify all existing projects

### Create a new project

```
POST /projects
```

* `shorthand` MUST be unique
* `shorthand` MUST be less than 31 bytes
* `name` MUST be less than 65536 bytes
* `description` MUST be less than 65536 bytes

#### Parameters

Name | Type | Description
-----|------|------------
name | string | The human-readable name of the project
shorthand| string | The shorthand identifier for archive filenames. Must be unique
description | string | A longer description of the project
status | string | The current status of the project. One of "active", "completed", "stalled" or "dropped"

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
shorthand | string | The unique project name shorthand, used in file names
status | string | The current status of the project
description | string | The description of the project

### Update project information

```
PUT /projects/{projectId}
```

* A project with id `projectId` MUST exist
* `shorthand` MUST be unique
* `shorthand` MUST be less than 31 bytes
* `name` MUST be less than 65536 bytes
* `description` MUST be less than 65536 bytes

#### Paramters

Name | Type | Description
-----|------|------------
projectId | integer | The unique identifier for the project
name | string | A new human-readable name for the project
shorthand | string | A new unique shorthand name, to use in filenames
status | string | The new status of the project
description | string | A new description for the project

#### Response

Name | Type | Description
-----|------|------------
success | bool | True if the update could take place, else false

### Delete a project

```
DELETE /projects/{projectId}
```

* A project with id `projectId` MUST exist
* There MUST be 0 associated releases

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

* A project with id `projectId` MUST exist

#### Parameters

Name | Type | Description
-----|------|------------
projectId | integer | The unique identifier for the project

#### Response

Name | Type | Description
-----|------|------------
releases | `[{"id": integer, "identifier": string, "scanlator": string, "status": string, "releasedOn": string, "version": integer}]` | An array of objects containing information about releases

### Create a new release

```
POST /projects/{projectId}/releases
```

#### Parameters

* A project with id `projectId` MUST exist
* `identifier` MUST be unique for that project
* `identifier` MUST be less than 11 bytes

Name | Type | Description
-----|------|------------
identifier | string | A unique identifier for the release
version | integer | The version of the release corresponding to that identifier
status | string | The release status, "released" or "draft"

#### Response

Name | Type | Description
-----|------|------------
success | bool | True if the release could be made, else false
id | integer | The newly created id of the release if successful

### Get information about a release

```
GET /projects/{projectId}/releases/{releaseId}
```

* A project with id `projectId` MUST exist
* A release with id `releaseId` MUST exist

#### Parameters

Name | Type | Description
-----|------|------------
projectId | integer | The unique id of the project under which the release was created
releaseId | integer | The unique id of the release

#### Response

Name | Type | Description
-----|------|------------
identifer | string | The unique identifier for the release
scanlator | string | The scanlator
version | integer | The release version number
status | string | The status of the release
releasedOn | string | The date that the release was made with its current status

### Update information about a release

* A project with id `projectId` MUST exist
* A release with id `releaseId` MUST exist
* `identifier` MUST be unique for that project
* `identifier` MUST be less than 11 bytes
* `version` MUST be greater than or equal to the previous version
* if `status` is changed from "draft" to "released", `version` MUST be greater than the previous version

```
PUT /projects/{projectId}/releases/{releaseId}
```

#### Parameters

Name | Type | Description
-----|------|------------
projectId | integer | The unique id of the project under which the release was created
releaseId | integer | The unique id of the release
identifier | string | The new unique identifier for the release
version | integer | The new version number for the release
status | string | The new status of the release

#### Response

Name | Type | Description
-----|------|------------
success | bool | True if the update could take place successfully, else false

### Delete a release

```
DELETE /projects/{projectId}/releases/{releaseId}
```

Deletes a release.
* A project with id `projectId` MUST exist
* A release with id `releaseId` MUST exist
* There MUST be 0 associated pages

#### Parameters

Name | Type | Description
-----|------|------------
projectId | integer | The unique id of the project under which the release was created
releaseId | integer | The unique id of the release

#### Response

Name | Type | Description
-----|------|------------
success | bool | True if the project could be deleted, else false

### Download an archive of a release

```
GET /projects/{projectId}/releases/{releaseId}/download/{archiveName}
```

* A project with id `projectId` MUST exist
* A release with id `releaseId` MUST exist
* The release MUST be in released state
* The archive MUST NOT contain any folders
* The archive MUST contain a credit page.
* Once released, the archive MUST remain unique.

The archive name follows the format {shorthand} - {identifier}[{version}][{groupName}].zip
Note that the square brackets are meant to be present in the actual URL. For example, one may request the following.

```
GET /projects/1/release/2/download/Robotics;Notes - Ch12[1][ims].zip
```

#### Parameters

Name | Type | Description
-----|------|------------
projectId | integer | The unique id of the project under which the release was created
releaseId | integer | The unique id of the release
archiveName | string | The name of the archive

#### Response

* Status 200: The zip file will be served directly
* Status 4xx: Invalid request, with a string error message
* Status 5xx: Server error, with a string error message

### Get a list of pages in a release

* A project with id `projectId` MUST exist
* A release with id `releaseId` MUST exist

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
pages | `[{"id": integer, "name": string, "createdAt": string, "mimeType": string}]`| An array of objects describing each of the pages part of the release

### Add a new page to a release

```
POST /projects/{projectId}/releases/{releaseId}/pages
```

* A project with id `projectId` MUST exist
* A release with id `releaseId` MUST exist
* `name` MUST end in either .png or .jpg
* `name` MUST be unique for that release
* `name` MUST be less than 256 bytes
* `data` MUST be a base64 encoded image of type matchign the extension in `name`

#### Parameters

Name | Type | Description
-----|------|------------
projectId | integer | The unique id of the project under which the release was created
releaseId | integer | The unique id of the release
name | string | The page number
data | string | The base64-encoded raw image data

#### Release

Name | Type | Description
-----|------|------------
success | bool | True if the page could be uploaded, else false e.g. if the release has a completed status
id | integer | The unique id of the newly created page

### Download a page

```
GET /projects/{projectId}/releases/{releaseId}/pages/{filename}
```

* A project with id `projectId` MUST exist
* A release with id `releaseId` MUST exist

#### Parameters

Name | Type | Description
-----|------|------------
projectId | integer | The unique id of the project under which the release was created
releaseId | integer | The unique id of the release
filename | string | The filename of the page

#### Response

* Status 200: The image file will be served directly
* Status 4xx: Invalid request, with a string error message
* Status 5xx: Server error, with a string error message

### Delete a page from a release

```
DELETE /projects/{projectId}/releases/{releaseId}/pages/{pageId}
```

* A project with id `projectId` MUST exist
* A release with id `releaseId` MUST exist

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
