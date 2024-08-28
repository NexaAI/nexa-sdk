import streamlit as st
from nexa.gguf import NexaImageInference

@st.cache_resource
def generate_ai_avatar():
    try:
        image_model = NexaImageInference(model_path="lcm-dreamshaper", local_path=None)

        images = image_model.txt2img(
            prompt="A girlfriend with long black hair",
            cfg_scale=image_model.params["guidance_scale"],
            width=image_model.params["width"],
            height=image_model.params["height"],
            sample_steps=image_model.params["num_inference_steps"],
            seed=image_model.params["random_seed"],
        )

        if images and len(images) > 0:
            avatar_path = "ai_avatar.png"
            images[0].save(avatar_path)
            return avatar_path
        else:
            st.error("No image was generated.")
            return None
    except Exception as e:
        st.error(f"Error generating AI avatar: {str(e)}")
        return None
