import sys
import os

base_path = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, base_path)
sys.path.insert(0, os.path.join(base_path, 'decoder'))
sys.path.insert(0, os.path.join(base_path, 'encoder'))