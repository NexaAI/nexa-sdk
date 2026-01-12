# Copyright 2024-2026 Nexa AI, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# !/usr/bin/env python3

css = """
.gallery-container {
    display: flex;
    flex-wrap: wrap;
    gap: 14px;
    justify-content: flex-start;
}

.card {
    position: relative;
    width: 150px;
    height: 150px;
    background-size: cover;
    background-position: center;
    border-radius: 12px;
    overflow: hidden;
    cursor: pointer;
    transition: transform 0.25s ease, box-shadow 0.25s ease;
    box-shadow: 0 2px 6px rgba(0,0,0,0.15);
}

.card:hover {
    transform: translateY(-4px) scale(1.03);
    box-shadow: 0 4px 12px rgba(0,0,0,0.25);
}

.video-card video {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.overlay-video-duration {
    position: absolute;
    top: 8px;
    left: 8px;
    display: flex;
    flex-direction: column;
    gap: 2px;
    background: rgba(0,0,0,0.45);
    color: #fff;
    padding: 4px 10px;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 400;
    text-align: right;
    max-width: 80%;
    white-space: nowrap;
    text-overflow: ellipsis;
    overflow: hidden;
}

.overlay-info {
    position: absolute;
    top: 8px;
    right: 8px;
    display: flex;
    flex-direction: column;
    gap: 2px;
    background: rgba(0,0,0,0.45);
    color: #fff;
    padding: 4px 10px;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 400;
    text-align: right;
    max-width: 80%;
    white-space: nowrap;
    text-overflow: ellipsis;
    overflow: hidden;
}

.info-line {
    line-height: 1.2;
}

#header-row {
    background-color: #F5F7F2;
    padding: 10px;
}

#input-row {
    align-items: center;
    border: 1px solid #e0e0e0;
    border-radius: 6px;
    padding: 8px;
}

#chat_input {
    border: none !important;
    outline: none !important; 
    box-shadow: none !important;  
    background: transparent; 
}

#chat_input textarea::-webkit-scrollbar { 
    display: none;
}

#model-dropdown {
    border: none !important;
    outline: none !important; 
    box-shadow: none !important;  
    background-color: transparent !important; 
}

#search-column {
    border: 1px solid #e0e0e0;
    border-radius: 6px;
    padding: 8px;
}

#send-btn {
    background-color: transparent;
    width: 40px;    
    height: 40px; 
    padding: 0;   
    margin: 0; 
}

.custom-btn {
    background-color: transparent !important;
    width: 30px !important;
    height: 30px !important;
    margin: 0 !important;
    padding: 0 !important;
}

.custom-btn2 {
    border-radius: 12px;
    border: 1px solid #454545;
    background-color: #FCFCFC !important;
    width: 30px !important;
    padding: 8px;
    font-size: 14px;
    font-style: normal;
    font-weight: 400;
    line-height: 20px; /* 142.857% */
    letter-spacing: 0.15px;
}

"""

