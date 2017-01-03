# IMangaScans API

## Convention

Unless specified otherwise, all responses will be served as JSON.  The common exception to this rule will be in cases where files are served rather than informational responses.

`GET` and `DELETE` requests will accept parameters in URL query string parameters.

`POST` and other types of requests will accept parameters as JSON in the body of the request.

JSON responses will always contain an `error` field.  If an error occurred in processing the request, it will contain a string describing the error. Otherwise, `error` will be `null`.

`POST`, `PUT` and `DELETE` requests will always require authentication via the `Auth-Token` header. The value set to the header must match the API's secret token.

## Notation

Types will be represented with the name of a type, such as `string`, `bool`, `integer`, `float` etc.

Compound types will be represented as `[<type>]` for arrays with values of all the same type, `[<type1>, <type2>, ... <typeN>]` in the case of an array with exactly N values whose types are predefined, and `{"key1": <type>, "key2": <type>}` in the case of objects. Objects may also have the type `{...}`, indicating that they can contain any combination of keys and values.

Any parameter that is optional will have a type prefixed with `optional`. For example, the type `optional integer` indicates that the field is not required but, if provided, must be an integer.

## Types

Below is a description of the various types

### Project

Name | Type | Description
-----|------|------------
id | integer | The project id
name | string | The descriptive name of the project
shorthand | string | The unique project name shorthand, used in file names
description | string | The description of the project
status | string | The current status of the project
createdAt | string | The date when the project was created

### Release

Name | Type | Description
-----|------|------------
id | integer | The release id
identifer | string | The unique identifier for the release
scanlator | string | The scanlator
version | integer | The release version number
status | string | The status of the release
releasedOn | string | The date that the release was made with its current status

### Page

Name | Type | Description
-----|------|------------
id | integer | The page id
name | string | Page filename
createdAt | string | The date when the page was created

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
error | string | Error string
result | Project[] | An array containing all existing projects

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
error | string | Error string
result | Project[] | An array containing the newly created project

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
error | string | Error string
result | Project[] | An array containing the project

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
error | string | Error string
result | Project[] | An array containing the updated project

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
error | string | Error string
result | Project[] | An array containing the deleted project

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
error | string | Error string
result | Release[] | An array containing all of the project's releases

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

#### Response

Name | Type | Description
-----|------|------------
error | string | Error string
result | Release[] | An array containing the newly created release

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
error | string | Error string
result | Release[] | An array containing the release

### Update information about a release

```
PUT /projects/{projectId}/releases/{releaseId}
```

* A project with id `projectId` MUST exist
* A release with id `releaseId` MUST exist
* `identifier` MUST be unique for that project
* `identifier` MUST be less than 11 bytes
* `version` MUST be greater than or equal to the previous version
* if `status` is changed from "draft" to "released", `version` MUST be greater than the previous version
* if current `status` is "released", the new status MUST be "draft"
* if new `status` is "released", there MUST be a credit page added to the release. A credit page shall be identified by a leading ! character in the filename.

#### Parameters

Name | Type | Description
-----|------|------------
projectId | integer | The unique id of the project under which the release was created
releaseId | integer | The unique id of the release
identifier | string | The new unique identifier for the release
version | integer | The new version number for the release
status | string | The new status of the release, one of "draft" or "released"

#### Response

Name | Type | Description
-----|------|------------
error | string | Error string
result | Release[] | An array containing the updated release

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
error | string | Error string
result | Release[] | An array containing the deleted release

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
* Status 4xx: Invalid request
* Status 5xx: Server error

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
error | string | Error string
result | Page[] | An array containing the pages for the release

### Add a new page to a release

```
POST /projects/{projectId}/releases/{releaseId}/pages
```

* A project with id `projectId` MUST exist
* A release with id `releaseId` MUST exist
* The release MUST be in draft state
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

#### Response

Name | Type | Description
-----|------|------------
error | string | Error string
result | Page[] | An array containing the newly created page

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
* Status 4xx: Invalid request
* Status 5xx: Server error

### Delete a page from a release

```
DELETE /projects/{projectId}/releases/{releaseId}/pages/{pageId}
```

* A project with id `projectId` MUST exist
* A release with id `releaseId` MUST exist
* The release MUST be in draft state

#### Parameters

Name | Type | Description
-----|------|------------
projectId | integer | The unique id of the project under which the release was created
releaseId | integer | The unique id of the release
pageId | integer | The unique identifier of the page to delete

#### Response

Name | Type | Description
-----|------|------------
error | string | Error string
result | Page[] | An array containing the deleted page

### Download a page thumbnail

```
GET /projects/{projectId}/releases/{releaseId}/thumbnails/{filename}
```

This route takes the same filename as the pages route, but returns a thumbnail
that is guaranteed to be at most 300 pixels wide or 200 pixels high.

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
* Status 4xx: Invalid request
* Status 5xx: Server error