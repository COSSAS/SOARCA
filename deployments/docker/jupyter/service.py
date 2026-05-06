import http.server
import socketserver
import subprocess
import sys

PORT = 7000


class CustomHandler(http.server.SimpleHTTPRequestHandler):
    def do_POST(self):
        # Get the notebook file from the message data
        input = self.rfile.read(int(self.headers.get('Content-Length')))
        # Run the notebook
        # The easiest way to run a one-shot notebook from a CLI is to convert it (to a notebook);
        # this allows us to set the --execute flag, which runs the notebook.
        # We handle input and output via stdin/stdout.
        try:
            res = subprocess.run(
                ["/home/jupy/lab/bin/jupyter", "nbconvert","--to","notebook","--execute","--stdin","--stdout"], 
                input = input,  capture_output=True, check=True)
            self.send_response_only(200)
            self.end_headers()
            self.wfile.write(res.stdout)
        except subprocess.CalledProcessError as e:
            self.send_response_only(500)
            self.end_headers()
            self.wfile.write(e.stderr)
        finally:
            sys.exit(0)

with socketserver.TCPServer(("", PORT), CustomHandler) as server:
    # We will only serve one request
    server.serve_forever()

