import os
from auth.session import Session

class Service:
    def handle(self, req):
        return validate(req)

def validate(req):
    return os.path.exists(req)
