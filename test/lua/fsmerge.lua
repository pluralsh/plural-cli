values = {}
valuesFiles = {}

files, err = fs.walk("test/lua/fleeta")
if err then
    error(err)
    return
end

table.sort(files, function(a, b)
    local aParts = utils.splitString(a, "/")
    local bParts = utils.splitString(b, "/")
    return #aParts < #bParts
end)

local workset = {}

for _, file in ipairs(files) do
    local f = fs.read(file)
    local asYaml = encoding.yamlDecode(f)
    local parent = {}
    local parts = utils.splitString(file, "/")
    local i = #parts - 1
    while i > 0 do
        local parentFile = table.concat(parts, "/", 1, i - 1) .. "/cluster.yaml"
        if  workset[parentFile] then
            parent = workset[parentFile]
            break
        end
        i = i - 1
    end
    workset[file] = utils.merge(parent, asYaml)
    if workset[file].cluster then
        values[workset[file].cluster] = workset[file]
    end
end
