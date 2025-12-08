from flask import Flask, render_template, request, jsonify, send_from_directory
from pathlib import Path
from datetime import datetime
from image_utils import image_to_base64
import uuid
import asyncio

import sys
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from function_call_util import run_agent

app = Flask(__name__)
app.config['MAX_CONTENT_LENGTH'] = 16 * 1024 * 1024  # 16MB max file size
app.config['UPLOAD_FOLDER'] = Path(__file__).resolve().parent / 'uploads'
app.config['UPLOAD_FOLDER'].mkdir(parents=True, exist_ok=True)

chat_history = []
processing_tasks = {}

def allowed_file(filename):
    """check allowed file types"""
    allowed_extensions = {'png', 'jpg', 'jpeg', 'gif', 'webp', 'bmp'}
    return '.' in filename and filename.rsplit('.', 1)[1].lower() in allowed_extensions

def process_message(text_content=None, image_path=None):
    """handle user message"""
    message = {
        'type': 'user',
        'timestamp': datetime.now().isoformat(),
        'text': text_content if text_content else None,
        'image': None
    }
    
    if image_path:
        try:
            message['image'] = image_to_base64(image_path)
        except Exception as e:
            return None, f"Image handle error: {str(e)}"
    
    chat_history.append(message)
    return message, None

def add_bot_response(response_type='text', content=None):
    bot_message = {
        'type': 'bot',
        'timestamp': datetime.now().isoformat(),
        'response_type': response_type,  # 'text', 'event'
        'content': content
    }
    chat_history.append(bot_message)
    return bot_message

@app.route('/')
def index():
    """chat page"""
    return render_template('chat.html')

@app.route('/api/send-message', methods=['POST'])
def send_message():
    """
    handle user message
    inputs: text and/or image
    return: task_id
    """
    try:
        text_content = request.form.get('message', '').strip()
        image_file = request.files.get('image')
        
        if not text_content and not image_file:
            return jsonify({'error': 'Please provide text or image'}), 400
        
        image_path = None
        if image_file and image_file.filename:
            if not allowed_file(image_file.filename):
                return jsonify({'error': 'not allowed file'}), 400
            
            # save uploaded image
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S_')
            filename = timestamp + image_file.filename
            image_path = app.config['UPLOAD_FOLDER'] / filename
            image_file.save(image_path)
        
        # handle user message
        user_message, error = process_message(text_content, image_path)
        if error:
            return jsonify({'error': error}), 400
        
        # create task id
        task_id = str(uuid.uuid4())
        # register task so client can poll /api/get-response/<task_id>
        processing_tasks[task_id] = {
            'status': 'pending',
            'user_text': text_content,
            'image_path': str(image_path) if image_path else None
        }
        
        return jsonify({
            'user_message': user_message,
            'task_id': task_id,
            'success': True
        })
    
    except Exception as e:
        return jsonify({'error': f'Serve error: {str(e)}'}), 500

@app.route('/api/get-response/<task_id>', methods=['GET'])
async def get_response(task_id):
    if task_id in processing_tasks:
        task = processing_tasks[task_id]
        if task['status'] == 'done':
            # clear task
            del processing_tasks[task_id]
            return jsonify(task['result'])
        elif task['status'] == 'processing':
            return jsonify({'status': 'processing'})
        elif task['status'] == 'pending':
            # process now (synchronous handling for simplicity)
            processing_tasks[task_id]['status'] = 'processing'
            try:
                
                
                response = await run_agent(
                    text=task['user_text'],
                    image=task['image_path']
                )
                
                import json
                data = json.loads(response)
                is_error = bool(data.get("isError"))
                content = data.get("content") or []
                bot_response = None
                if is_error:
                    bot_response = add_bot_response(
                        response_type='text',
                        content=content[0]["text"]
                    )
                else:
                    text = json.loads(content[0]["text"])
                    event = text["event"]
                    print(f"[info] Event detected: {event}")
                    summary=event.get("summary", "No Title")
                    date = "N/A"
                    if event.get("start"):
                        start_time = event["start"].get("dateTime", "N/A")
                        date = start_time.split("T")[0]
                        start_time = start_time.split("T")[1] if "T" in start_time else "N/A"
                    else:
                        start_time = "N/A"
                    if event.get("end"):
                        end_time = event["end"].get("dateTime", "N/A")
                        end_time = end_time.split("T")[1] if "T" in end_time else "N/A"
                    else:
                        end_time = "N/A"
                    venue = event.get("location", "N/A")
                    description = event.get("description", summary)
                    address = event.get("address", "N/A")
                    htmlLink = event.get("htmlLink", "")
                    bot_response = add_bot_response(
                        response_type='event',
                        content={
                            'event_name': summary,
                            'date': date,
                            'start_time': start_time,
                            'end_time': end_time,
                            'venue': venue,
                            'address': address,
                            'description': description,
                            'htmlLink': htmlLink
                        }
                    )
                result = {'status': 'done', 'bot_response': bot_response}
                processing_tasks[task_id]['status'] = 'done'
                processing_tasks[task_id]['result'] = result
                # return and let client clear on next poll
                return jsonify(result)
            except Exception as e:
                bot_response = add_bot_response(
                    response_type='text',
                    content=str(e)
                )
                result = {'status': 'done', 'bot_response': bot_response}
                processing_tasks[task_id]['status'] = 'done'
                processing_tasks[task_id]['result'] = result
                return jsonify(result)
    
    return jsonify({
        'status': 'done'
    })

@app.route('/api/chat-history', methods=['GET'])
def get_chat_history():
    return jsonify(chat_history)

@app.route('/api/clear-history', methods=['POST'])
def clear_history():
    global chat_history
    chat_history = []
    return jsonify({'success': True})

@app.route('/uploads/<path:filename>')
def serve_upload(filename):
    return send_from_directory(app.config['UPLOAD_FOLDER'], filename)

if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0', port=3000)
