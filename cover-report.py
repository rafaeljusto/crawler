#!/usr/bin/env python

# Copyright 2014 Rafael Dantas Justo. All rights reserved.
# Use of this source code is governed by a GPL
# license that can be found in the LICENSE file.

import os
import subprocess
import sys

def initialChecks():
  if "GOPATH" not in os.environ:
    print("Need to set GOPATH")
    sys.exit(1)

def findPath():
  goPath = os.environ["GOPATH"]
  goPathParts = goPath.split(";")
  for goPathPart in goPathParts:
    projectPath = os.path.join(goPathPart, "src", "github.com",
      "rafaeljusto", "crawler")
    if os.path.exists(projectPath):
      return projectPath

  return ""

def changePath():
  projectPath = findPath()
  if len(projectPath) == 0:
    print("Project not found")
    sys.exit(1)

  os.chdir(projectPath)

def runCoverReport():
  success = True

  try:
    subprocess.check_call(["go", "install"])
    subprocess.check_call(["go", "test", "-coverprofile=cover-profile.out", "-cover"])
    subprocess.check_call(["go", "tool", "cover", "-html=cover-profile.out"])
  except subprocess.CalledProcessError:
    success = False

  # Remove the temporary file created for the
  # covering reports
  try:
    os.remove("cover-profile.out")
  except OSError:
    pass

  if not success:
    print("Errors during the unit test execution")
    sys.exit(1)

###################################################################

if __name__ == "__main__":
  try:
    initialChecks()
    changePath()
    runCoverReport()
  except KeyboardInterrupt:
    sys.exit(1)
