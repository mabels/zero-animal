import subprocess
import json
import sys

result = subprocess.run(['docker', 'manifest', 'inspect', 'alpine:3'], stdout=subprocess.PIPE)
obj = json.loads(result.stdout.decode('utf-8'))
#print(obj)

platformItems = sys.argv[-1].split('/')
os=platformItems[0]
architecture=platformItems[1]
if len(platformItems) > 2:
    variant=platformItems[2]
else:
    variant=None

#      {
#         "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
#         "size": 528,
#         "digest": "sha256:e7d88de73db3d3fd9b2d63aa7f447a10fd0220b7cbf39803c803f2af9ba256b3",
#         "platform": {
#            "architecture": "amd64",
#            "os": "linux"
#         }
#      },
def match(o):
    platform = o["platform"]
    #print(json.dumps(o))
    if not ("architecture" in platform and platform["architecture"] == architecture):
        return False
    if not ("os" in platform and platform["os"] == os):
        return False
    if variant is not None:
        if not ("variant" in platform and variant == platform["variant"]):
          return False
    return True
manifests = list(filter(match, obj["manifests"]))
if len(manifests):
    print(json.dumps(manifests[0]['digest']))
else:
    os.exit(1)




