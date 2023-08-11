#!/usr/bin/env python3
import yaml
import os

def create_structure(structure, parent_path="."):
    for item in structure:
        if "dir" in item:
            dir_path = os.path.join(parent_path, item["dir"])
            os.makedirs(dir_path, exist_ok=True)
            if "children" in item:
                create_structure(item["children"], dir_path)
        elif "file" in item:
            file_path = os.path.join(parent_path, item["file"])
            open(file_path, 'a').close()

if __name__ == "__main__":
    with open(".project.yaml", 'r') as file:
        project = yaml.safe_load(file)
        create_structure(project["structure"])
