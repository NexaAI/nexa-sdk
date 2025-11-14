# !/usr/bin/env python3

import html
import os
import gradio as gr
from typing import List

# Import ViewModel and related classes/constants
from viewmodel import (
    ViewModel, SearchResult, metrics
)

from style import css


vm = ViewModel()

##############################
# events handlers
##############################

def on_files_chage(files):
    """Handle file upload event."""
    if files is None or len(files) == 0:
        vm.files = []
    else:
        vm.files = [file.name for file in files]

def on_search_click(query: str):
    """Handle search button click event."""
    try:
        (videos, images, search_time) = vm.search(query, alogrithm="embed-neural")
        return (
                gr.update(visible=False),
                gr.update(visible=False),
                gr.update(visible=True),
                render_items("Images", images),
                gr.update(label=f"Images({len(images)})"),
                gr.update(visible=True, value=f"{search_time:.3f}s"),
            )
    except Exception as e:
        return (
                gr.update(visible=True, value=f"<div align='center'><h3>{e}</h3></div>"),
                gr.update(visible=False),
                gr.update(visible=False),
                render_items("Images", []),
                gr.update(label=f"Images(0)"),
                gr.update(visible=False, value=""),
            )

def on_load_click():
    """Handle load model button click event."""
    vm.load_model("EmbedNeuralText/EmbedNeuralImage")
    return gr.update(value="Load Model", interactive=True)

def on_unload_click():
    vm.unload_model()
    return gr.update(interactive=True)


def on_download_click():
    """Handle download button click event."""
    try:
        vm.download("EmbedNeuralText/EmbedNeuralImage")
    except Exception as e:
        print(f"Error during model download: {e}")
    return gr.update(value="Download", interactive=True)

def local_to_url(path):
    if os.path.exists(path):
        return f"/gradio_api/file={os.path.abspath(path)}"
    return path 

def get_image_base64(url: str):
    with open(url, 'rb') as f:
        import base64
        data = base64.b64encode(f.read()).decode()
        return f"data:image/png;base64,{data}"

# Build Gradio UI
def render_items(tab: str, items: List[SearchResult]):
    # Renders search results as HTML cards for images/videos
    if(items is None or len(items) == 0): 
        return f"<div class='gallery-container'>{tab} not found</div>"
    html_items = ""
    for item in items:
        # Extract time and duration information
        start = html.escape(f"{item.start:.1f}")
        end = html.escape(f"{item.end:.1f}")
        duration = f"{start}s - {end}s"
        base64_data = get_image_base64(item.url)
        is_image = item.is_image
        if tab == "Images":
            # Render image card without score overlay
            html_items += f"""
            <div class="card" style="background-image: url('{base64_data}');">
            </div>
            """
        else:
            if is_image:
                # Render image card with timestamp for video frames
                html_items += f"""
                <div class="card" style="background-image: url('{base64_data}');">
                    <div class="overlay-video-duration">{start}s</div>
                </div>
                """
            else:
                # Render video card with duration
                url = local_to_url(item.url)
                html_items += f"""
                <div class="card video-card">
                    <video src="{url}" control></video>
                    <div class="overlay-video-duration">{duration}</div>
                </div>
                """
    return f"<div class='gallery-container'>{html_items}</div>"

def load_images():
    return render_items("Images")

def load_videos():
    return render_items("Videos")


# main interface
with gr.Blocks(title="Image/Video Search", fill_height=True, css=css) as demo:
    with gr.Row(elem_id='header-row'):
        gr.HTML("")
        download_btn = gr.Button("Download", min_width=150, scale=0, elem_classes='custom-btn2')
        load_btn = gr.Button("Load Model", min_width=150, scale=0, elem_classes='custom-btn2')
        unload_btn = gr.Button("Unload Model", min_width=150,scale=0, elem_classes='custom-btn2')
  
    with gr.Row():
        with gr.Column():
            uploader = gr.Files(
                label="Upload files (mp4, png, jpg, jpeg)",
                file_types=['.mp4', '.png', '.jpg', '.jpeg'],
                file_count="multiple",
                height=500,
            )
            
        with gr.Column(scale=8, elem_id='search-column'):
            placeholder = gr.Markdown("<div align='center'><h3>Please import files before starting your search.</h3></div>", visible=True, height=500)
            searching_box = gr.Markdown("### 🔍 Searching...", visible=False, height=500)
            
            # Results display (images only)
            with gr.Tabs(visible=False) as result_tabs:
                with gr.Tab(f"Images") as img_tab:
                    images_tab = gr.HTML(min_height=500, max_height=500)

            with gr.Row(elem_id='input-row'):
                chat_input = gr.Textbox(show_label=False, container=False, placeholder="Search item...", lines=1, elem_id="chat_input")
                send_btn = gr.Button(value="", icon="images/button-bg.png", elem_classes="custom-btn", min_width=30, scale=0)
        
        with gr.Column(min_width=200):
            with gr.Group():
                top_k = gr.Number(label="Top-K", value=2, step=1, minimum=1, interactive=True)
                matric = gr.Dropdown(metrics, label="Metric", value=metrics[0], container=True, interactive=True)
            
            search_time = gr.Textbox(label="Search Time (s)", value="", interactive=False, lines=1, visible=False)
      
    
    top_k.change(
        fn=lambda value: vm.update_top_k(int(value)),
        inputs=[top_k],
        outputs=[]
    )

    matric.change(
        fn=lambda value: vm.update_metric(value),
        inputs=[matric],
        outputs=[]
    )
    
    # File upload handler
    uploader.change(fn=on_files_chage, inputs=[uploader], outputs=[])
    
    # Chat input submit handler
    chat_input.submit(
        fn=lambda: (gr.update(visible=False), gr.update(visible=True), gr.update(visible=False), gr.update(visible=False)),
        outputs=[placeholder, searching_box, result_tabs, search_time],
    ).then(
        fn=on_search_click,
        inputs=[chat_input],
        outputs=[placeholder, searching_box, result_tabs, images_tab, img_tab, search_time]
    )
    
    # Send button click handler
    send_btn.click(
        fn=lambda: (gr.update(visible=False), gr.update(visible=True), gr.update(visible=False), gr.update(visible=False)),
        outputs=[placeholder, searching_box, result_tabs, search_time],
    ).then(
        fn=on_search_click,
        inputs=[chat_input],
        outputs=[placeholder, searching_box, result_tabs, images_tab, img_tab, search_time]
    )
    
    # Download button click handler
    download_btn.click(
        fn=lambda: gr.update(value="Downloading...", interactive=False),
        outputs=[download_btn]
    ).then(
        fn=on_download_click,
        inputs=[],
        outputs=[download_btn]
    )
    
    # Load button click handler
    load_btn.click(
        fn=lambda: gr.update(value="Loading...", interactive=False),
         outputs=[load_btn]
    ).then(
        fn=on_load_click,
        inputs=[],
        outputs=[load_btn]
    )
    
    # Unload button click handler
    unload_btn.click(
        fn=lambda: gr.update(interactive=False),
         outputs=[unload_btn]
    ).then(
        fn=on_unload_click,
        inputs=[],
        outputs=[unload_btn]
    )
    
if __name__ == "__main__":
	demo.launch(allowed_paths=["./images/button-bg.png"])

