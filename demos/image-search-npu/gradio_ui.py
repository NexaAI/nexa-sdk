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
        return gr.update(interactive=False)
    else:
        vm.files = [file.name for file in files]
        return gr.update(interactive=True)

def on_index_click():
    """Handle index button click event."""
    try:
        vm.index_files()
        return (
            gr.update(value="Index", interactive=True),
            gr.update(visible=False, value=100)
        )
    except Exception as e:
        print(f"Error during indexing: {e}")
        return (
            gr.update(value="Index", interactive=True),
            gr.update(visible=False, value=0)
        )

def on_search_click(query: str):
    """Handle search button click event."""
    try:
        (_, images, search_time) = vm.search(query)
        return (
                gr.update(visible=False),
                gr.update(visible=False),
                gr.update(visible=True),
                render_items(images),
                gr.update(label=f"Images({len(images)})"),
                gr.update(visible=True, value=f"{search_time:.3f}s"),
            )
    except Exception as e:
        return (
                gr.update(visible=True, value=f"<div align='center'><h3>{e}</h3></div>"),
                gr.update(visible=False),
                gr.update(visible=False),
                render_items([]),
                gr.update(label=f"Images(0)"),
                gr.update(visible=False, value=""),
            )

def get_image_base64(url: str):
    with open(url, 'rb') as f:
        import base64
        data = base64.b64encode(f.read()).decode()
        return f"data:image/png;base64,{data}"

# Build Gradio UI
def render_items(items: List[SearchResult]):
    # Renders search results as HTML cards for images
    if(items is None or len(items) == 0): 
        return f"<div class='gallery-container'>No images found</div>"
    html_items = ""
    for item in items:
        # Get base64 encoded image data
        base64_data = get_image_base64(item.url)
        # Render image card without score overlay
        html_items += f"""
        <div class="card" style="background-image: url('{base64_data}');">
        </div>
        """
    return f"<div class='gallery-container'>{html_items}</div>"


# main interface
with gr.Blocks(title="Image Search", fill_height=True, css=css) as demo:
    with gr.Row():
        with gr.Column():
            uploader = gr.Files(
                label="Upload images (png, jpg, jpeg)",
                file_types=['.png', '.jpg', '.jpeg'],
                file_count="multiple",
                height=500,
            )
            index_btn = gr.Button("Index", elem_classes="custom-btn2", min_width=400, interactive=False)
            index_progress = gr.Slider(minimum=0, maximum=100, interactive=False, label="Indexing progress", value=0, visible=False)
            
        with gr.Column(scale=8, elem_id='search-column'):
            placeholder = gr.Markdown("<div align='center'><h3>Please import files and click Index before starting your search.</h3></div>", visible=True, height=500)
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
    uploader.change(fn=on_files_chage, inputs=[uploader], outputs=[index_btn])
    
    # Index button click handler
    index_btn.click(
        fn=lambda: (gr.update(value="Indexing...", interactive=False), gr.update(visible=True, value=0)),
        outputs=[index_btn, index_progress]
    ).then(
        fn=on_index_click,
        inputs=[],
        outputs=[index_btn, index_progress]
    )
    
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
    
if __name__ == "__main__":
	import os
	# Enable hot reload by setting environment variable or using watch parameter
	# You can also run: python gradio_ui.py --reload (if supported)
	demo.launch(allowed_paths=["./images/button-bg.png"])

