# /// script
# dependencies = [
#   "pyautogui",
# ]
# ///


import os
import subprocess
import time

import pyautogui


def do(inputs: list):
    for entry in inputs:
        time.sleep(0.1)
        # Lists are keys, tuples are written as text, ints are delays
        if isinstance(entry, list):
            if len(entry) == 1:
                pyautogui.press(entry[0])
            else:
                # press the key entry[0] entry[1] times with a delay of entry[2] ms
                for _ in range(entry[1]):
                    pyautogui.press(entry[0])
                    if entry[1] != 1:
                        time.sleep(entry[2] / 1000)

        elif isinstance(entry, int):
            time.sleep(entry / 1000)
        else:
            pyautogui.write(entry, interval=0.07)


# Start new terminal by executing stui in alacritty as root, do not block
# We need root user here in order to be able to use `scontrol update` commands
homedir = os.path.expanduser("~")
pipe = subprocess.Popen(
    ["sudo", "alacritty", "-e", f"{homedir}/go/bin/stui"],
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE,
)

input("ok to proceed?")
print("proceeding in 10 seconds...")

for i in range(10, 0, -1):
    print(i)
    time.sleep(1)


## NODES
do(
    [
        ["1"],  # choose nodes view
        ["j", 3, 100],  # go down
        ["p", 1, 100],  # choose partition
        [
            "down",
            2,
            100,
        ],  # go down in the partition list - this puts you in 'chemistry'
        ["enter"],  # select the partition
        ["s", 1, 100],  # focus node state
        ("idl"),  # search for 'idl' for idle
        ["enter"],  # select idle
        ["/"],  # open search view
        ("linux..3"),  # search for linux..3
        ["enter"],  # exit search view
        [" "],  # select
        ["j"],  # down
        [" "],  # select a node
        ["j"],
        [" "],  # down one, select a node
        ["enter"],  # view details
        300,
        ["esc"],  #  exit
        ["c"],  # enter command mode
        150,
        ("state=DRAIN"),  # set state to drain
        75,
        ["enter"],  # execute, this will fail
        200,
        (" reason='maintenance'"),  # set state to drain
        ["enter"],  # execute, this will succed
        250,
        ["esc"],  # exit
    ]
)

time.sleep(2)

## JOBS SCREEN
do(
    [
        ["2"],  # choose jobs view
        ["j", 3, 100],  # go down
        ["p", 1, 100],  # choose partition
        ["down", 2, 100],  # go up in the partition list
        ["enter"],  # select the partition, this puts you in 'biology'
        ["s"],  # focus state
        ("run"),  # search for RUNNING state
        ["enter"],  # confirm
        ["up", 1, 100],  # go up
        ["/"],  # open search view
        ["backspace", 8, 15],  # select the partition
        ("spare"),  # search for 'spare' to find QOS
        ["enter"],  # exit search view
        [" "],  # select
        ["j"],
        [" "],
        ["j"],
        [" "],
        ["enter"],  # view details
        500,
        ["esc"],  # exit
        ["c"],  # enter command mode
        ("timelimit=1-12:00:00"),
        ["enter"],  # execute
        250,
        ["esc"],  # exit
        ["esc"],  # exit
        ["esc"],  # exit
    ]
)


## SDIAG
do(
    [
        ["3"],  # choose sdiag view
        100,
        ["down", 10, 10],  # go down in the partition list
    ]
)


## SACCT
do(
    [
        ["4"],
        ["esc"],  # exit search view, clear out search
        50,
        ["down", 5, 25],
        ["e"],
        ["down", 10, 25],  # this gives you 'Transaction'
        ["enter"],  # confirm entity selection
        ["/"],  # open search view
        ("john"),
        ["enter"],  # exit search view
        ["down", 3, 25],
    ]
)
