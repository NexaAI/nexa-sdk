import streamlit as st
from PIL import Image
from streamlit_modal import Modal
import os
import shutil
from utils.initialize import initialize_or_update_system_message


def initialize_temp_customization():
    if "temp_customization" not in st.session_state:
        st.session_state.temp_customization = {
            "name": st.session_state.get("soulmate_name", "Yuzuki"),
            "gender": st.session_state.get("soulmate_gender", "Female"),
            "custom_instructions": st.session_state.get("custom_instructions", ""),
        }


def customize_avatar():
    uploaded_file = st.file_uploader("Choose an image...", type=["jpg", "jpeg", "png"])
    if uploaded_file is not None:
        image = Image.open(uploaded_file)
        uploaded_avatar_path = "uploaded_avatar.png"
        image.save(uploaded_avatar_path)
        st.session_state.uploaded_avatar = uploaded_avatar_path
        st.image(image, caption="Preview", use_column_width=False)

def customization_form():
    initialize_temp_customization()

    st.subheader("1. Change Avatar")
    customize_avatar()
    st.markdown("<br>", unsafe_allow_html=True)

    st.subheader("2. Customize Character")
    st.session_state.temp_customization["name"] = st.text_input(
        "Name:", value=st.session_state.temp_customization["name"]
    )
    st.markdown("<br>", unsafe_allow_html=True)

    st.session_state.temp_customization["gender"] = st.radio(
        "Gender:",
        ["Female", "Male"],
        index=0 if st.session_state.temp_customization["gender"] == "Female" else 1,
    )
    st.markdown("<br>", unsafe_allow_html=True)

    st.session_state.temp_customization["custom_instructions"] = st.text_area(
        "Custom Instructions:",
        value=st.session_state.temp_customization["custom_instructions"],
    )
    st.markdown("<br>", unsafe_allow_html=True)

    # CSS to center the buttons
    st.markdown(
        """
        <style>
        .stButton > button {
            display: block;
            margin: 0 auto;
        }
        </style>
    """,
        unsafe_allow_html=True,
    )

    col1, col2 = st.columns([1, 1])

    with col1:
        if st.button("Apply Changes"):
            apply_changes()
            st.session_state.modal_open = False
            st.rerun()

    with col2:
        if st.button("Cancel"):
            close_modal()


def close_modal():
    st.session_state.modal_open = False
    if "temp_customization" in st.session_state:
        del st.session_state.temp_customization
    if "uploaded_avatar" in st.session_state:
        if os.path.exists(st.session_state.uploaded_avatar):
            os.remove(st.session_state.uploaded_avatar)
        del st.session_state.uploaded_avatar
    st.rerun()


def apply_changes():
    # update avatar:
    if "uploaded_avatar" in st.session_state:
        permanent_avatar_path = "ai_avatar.png"
        shutil.copy(st.session_state.uploaded_avatar, permanent_avatar_path)
        st.session_state.ai_avatar = permanent_avatar_path
        os.remove(st.session_state.uploaded_avatar)
        del st.session_state.uploaded_avatar

    # if no new avatar was uploaded:
    if "ai_avatar" not in st.session_state:
        st.session_state.ai_avatar = "ai_avatar.png"

    # update other customization options:
    st.session_state.soulmate_name = st.session_state.temp_customization["name"]
    st.session_state.soulmate_gender = st.session_state.temp_customization["gender"]
    st.session_state.custom_instructions = st.session_state.temp_customization[
        "custom_instructions"
    ]

    # clear previous chat messages:
    st.session_state.messages = []

    # update system prompt with new customization:
    initialize_or_update_system_message()
    # new_prompt = f"""
    # You are {st.session_state.soulmate_name}, my perfect {st.session_state.soulmate_gender.lower()} soulmate.
    # {st.session_state.custom_instructions}
    # Start by introducing yourself briefly. You will respond in a concise way.
    # """
    # st.session_state.messages = [{"role": "system", "content": new_prompt.strip()}]

    # set flag to indicate customization was applied:
    st.session_state.customization_applied = True
    # clean up temporary customization data:
    del st.session_state.temp_customization

    st.session_state.modal_open = False

    if "intro_sent" in st.session_state:
        del st.session_state["intro_sent"]

    # force a rerun to update the UI:
    st.rerun()


def open_customization_modal():
    if "modal_open" not in st.session_state:
        st.session_state.modal_open = False

    if st.button("Customize Character"):
        st.session_state.modal_open = True
        initialize_temp_customization()

    if st.session_state.modal_open:
        modal = Modal("Customize Your Soulmate", key="customize_modal")
        with modal.container():
            customization_form()
