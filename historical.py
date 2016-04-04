## TODO: Allow arbitrary ordering
## TODO: Clean up - think better about selections, commands displayed, etc

import re
from subprocess import Popen, PIPE, STDOUT
import os
import stat

OKBLUE = '\033[94m'
OKGREEN = '\033[92m'
WARNING = '\033[93m'
FAIL = '\033[91m'
ENDC = '\033[0m'
FINISHED = False

BASH_SCRIPT_TEMPLATE = "#!/bin/bash\n"

def obtain_commands():
  shell_command = 'bash -i -c "history -r; history"'
  event = Popen(shell_command, shell=True, stdin=PIPE, stdout=PIPE, 
      stderr=STDOUT)

  output = event.communicate()
  history_line = r"(\d+) (.*)$"
  commands = [re.match(history_line, hl.strip()).groups() for hl in [c for c in output[0].split('\n')][1:-1]]
  commands = [c.strip() for l, c in commands[-100:]]
  return commands

def display_commands(commands):
  for i, c in enumerate(commands):
    print("%s: %s" % (i, c))

def update_commands_with_selection(selections, commands):
  numerical_selections = [int(s) for s in selections]
  for i, c in enumerate(commands):
    if i in numerical_selections:
      s = OKGREEN + "%s: %s" + ENDC
      print(s % (i, c))
    else:
      print("%s: %s" % (i, c))

def sanitize_script_name(script_name):
  return script_name

def create_script_from_selections(script_name, selected_commands, commands):
  name = script_name + '.sh'
  with open(name, 'w') as f:
    f.write(BASH_SCRIPT_TEMPLATE)
    numerical_selections = [int(s) for s in selected_commands]
    for i, c in enumerate(commands):
      if i in numerical_selections:
        f.write('%s\n' % c)
  os.chmod(name, os.stat(name).st_mode | stat.S_IEXEC)

SELECTED = False
commands = None
selected_commands = None

while not FINISHED:
  if not SELECTED:
    print("Enter y to select from the most recent 100 history commands to construct a script or n to exit")
    ready = raw_input('Ready?: y or n: ')
    if ready == 'n':
      FINISHED = True
    elif ready == 'y':
      commands = obtain_commands()
      display_commands(commands)
      selected_commands = raw_input('Selected Commands: ')
      selected_commands = selected_commands.split(',')
      update_commands_with_selection(selected_commands, commands)
      SELECTED = True
  else:
    finished = raw_input('Ready to output script? y or n: ')
    if ready == 'y':
      script_name = raw_input('Name...that..script!: ')
      script_name = sanitize_script_name(script_name)
      print('Creating script %s' % script_name)
      create_script_from_selections(script_name, selected_commands, commands)
      FINISHED = True
    
