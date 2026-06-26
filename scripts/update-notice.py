#!/bin/env python3

# Usage: go-licenses report . | ./scripts/update-notice.py > NOTICE

from sys import stdin

NOTICE_START = """NOTICE

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0).

It also makes use of third-party open source libraries. These dependencies are included under their respective licenses and are not covered by this project's AGPL license. All third-party code remains under its original license.

Below is a list of libraries and their licenses:
"""

NOTICE_END = "Please refer to each library’s repository for its full license text and source code."

licenses = {}

for line in stdin:
    values = line.strip().split(',')
    name = values[0]
    link = values[1]
    license_type = values[2]

    if "4rneee/noah" in name:
        continue

    t = (name, link)

    if license_type in licenses:
        licenses[license_type].append(t)
    else:
        licenses[license_type] = [t]

print(NOTICE_START)

for license_type, modules in licenses.items():
    print(f"{license_type} License:")
    for (name, link) in modules:
        print(f"- {name} - {link}")
    print("")

print(NOTICE_END)
