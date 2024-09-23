import streamlit as st
from nexa.gguf import NexaTextInference
import json
import logging
from utils.validate_json import validate_json

logger = logging.getLogger(__name__)


with open('victim_json_template_flat.json', 'r') as f:
    victim_info = json.load(f)

initial_prompt = "Hi, I am SafeGuardianAI. I am here to provide you with support and inform rescue teams of your vital status. Can you speak?"

system_instructions = f'''You are a post-disaster bot. Help victims while collecting valuable data for intervention teams. Your aim is to complete this template: {victim_info}. Only return JSON output when calling function.'''

def initialize_chat():
    if "messages" not in st.session_state or not st.session_state.messages:
        #print('hello')
        st.write(initial_prompt)
    if "victim_number" not in st.session_state:
        st.session_state['victim_number'] = set_key(st.session_state['victim_info'])
    if "json_response" not in st.session_state:
        st.session_state['json_response'] = {}


        
@st.cache_resource
def load_model(model_path):
    st.session_state.messages = []
    nexa_model = NexaTextInference(
        model_path=model_path,
        local_path=None,
        temperature=0.9,
        max_new_tokens=1024,
        top_k=50,
        top_p=1.0,
    )
    return nexa_model


import firebase_admin
from firebase_admin import credentials
from firebase_admin import db
import json
import os
import datetime
time = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")

cred = credentials.Certificate("disasterrescueai-firebase-adminsdk.json")

try:
    firebase_admin.initialize_app(credential=cred,options={'databaseURL': 'https://disasterrescueai-default-rtdb.firebaseio.com'}, name='RescueTeam_RealTimeDatabase')
except:
    pass

db = firebase_admin.db

def set_key(json_data):
    ref = db.reference('rescue_team_dataset', app=firebase_admin.get_app(name='RescueTeam_RealTimeDatabase'))
    # Generate a unique key
    new_key = ref.push().key
    # Add the data with the unique key
    ref.child(new_key).set(json_data)
    return new_key


def update_time_and_status(ref, victim_id, time, rescue_status, emergency_status):
    ref.child(victim_id).update({'last_updated': time, 'rescue_status': rescue_status, 'emergency_status': emergency_status})

def update_(victim_id, json_data):
    # Check if key exists in the database
    ref = db.reference(f'rescue_team_dataset/', app=firebase_admin.get_app(name='RescueTeam_RealTimeDatabase'))
    # try to get emergency_status
    ref.child(victim_id).update(json_data)
    try:
        update_time_and_status(ref, victim_id, time, json_data.get('rescue_status'), json_data.get('emergency_status'))
    except:
        update_time_and_status(ref, victim_id, time, 'pending', 'low_priority')


# Function to display victim information
def display_victim_info() -> None:
    st.write("**Parsed Information:**")
    st.json(st.session_state.victim_info)
    # Send data to Firebase
    try:
        update_(
            victim_id=st.session_state['victim_number'],
            json_data=st.session_state['victim_info']
        )
        current_time = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        st.success(
            f"{current_time}\nID: {st.session_state['victim_number']}\nYour data has been sent to the Rescue Team."
        )
    except Exception as e:
        logger.error(f"Error sending data to Firebase: {e}")
        st.warning("Error sending data to the Rescue Team.")
