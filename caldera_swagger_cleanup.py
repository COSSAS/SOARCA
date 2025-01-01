import json


with open('caldera_api_raw.json', 'r') as fd:
    data = json.load(fd)


for p in data['paths'].values():
    for m in p.values():
        # Ensure everything has at least one possible response
        if not m['responses']:
            m['responses'] = {'200': {'description': "OK"}}

        if m['parameters']:
            for par in m['parameters']:
                if 'schema' in par and '$ref' not in par['schema']:
                    for k, v in par['schema'].items():
                        par[k] = v
                    del par['schema']
                
                if 'operation_id' in par:
                    del par['operation_id']

                if 'access' in par:
                    del par['access']

                if 'output' in par:
                    del par['output']

                if 'type' in par and par['type'] == 'file' and 'consumes' not in m:
                    m['consumes'] = ["multipart/form-data"]
                
                if 'required' not in par:
                    par['required'] = True

for d in data['definitions'].values():
    for p in d['properties'].values():
        if 'format' in p and not p['format']:
            del p['format']
        if 'pattern' in p and not p['pattern']:
            del p['pattern']


# When POSTing a new operation, Operation.Objective is an empty string while the spec
# specifies it to always be an Operation object. Patch this away.
del data['definitions']['Operation']['properties']['objective']


with open('caldera_api.json', 'w') as fd:
    json.dump(data, fd)
