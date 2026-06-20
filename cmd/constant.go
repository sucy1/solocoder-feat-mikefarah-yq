package cmd

var unwrapScalarFlag = newUnwrapFlag()

var printNodeInfo = false

var unwrapScalar = false

var writeInplace = false
var outputToJSON = false

var outputFormat = ""

var inputFormat = ""

var exitStatus = false
var indent = 2
var noDocSeparators = false
var nullInput = false
var nulSepOutput = false
var verbose = false
var version = false
var prettyPrint = false

var forceColor = false
var forceNoColor = false
var colorsEnabled = false

// can be either "" (off), "extract" or "process"
var frontMatter = ""

var splitFileExp = ""
var splitFileExpFile = ""

var completedSuccessfully = false

var forceExpression = ""

var expressionFile = ""

var sortByField = ""
var sortByReverseField = ""

var toYaml = false
var fromYaml = false
var toJson = false
var fromJson = false
var toXml = false
var fromXml = false
var toToml = false
var fromToml = false

var mergeAll = false
var mergeStrategy = "overwrite"

var noBackup = false

var typeGuard = ""
