
<h2>🌟 Overview</h2>

<p><b>SafeGuardianAI</b> is an innovative emergency response platform that harnesses the power of artificial intelligence to provide critical support during catastrophic events. 
By seamlessly integrating real-time data analysis, offline capabilities, and community-driven support, SafeGuardianAI bridges the gap between individuals, communities, and emergency services, ensuring a more coordinated and effective response to disasters.</p>

<h3>NexaAI-SDK</h3>
<li><b> Conversational Bot: </b> Llama3-8B-Lexi-Uncensored</li>
<li><b> Data-Parsing Bot : </b> llama-3.1-8b-instant</li>
<li><b> Automatic Speech recognition</b> : Faster-Whisper-Tiny</li>
<li><b> Text-To-Speech</b> : OpenAI </li>



<h2>🛠️ Key Features</h2><ul>
  
  <li>💬 <strong>AI-Powered Chat Interface</strong>: Interact with GuardianAI for personalized emergency assistance and real-time guidance.</li>
  <li>🎙️ <strong>Multi-Modal Activation</strong>: Trigger emergency mode via text, voice commands, or automatic motion detection for hands-free operation.</li>
  <li>👤 <strong>Smart User Profiles</strong>: Store vital information securely for quick access during emergencies, including medical history and emergency contacts.</li>
  <li>📱 <strong>Offline Functionality</strong>: Access critical guides and emergency procedures without an internet connection, ensuring help is always available.</li>
  <li>🧠 <strong>Real-time Situation Assessment</strong>: Receive tailored responses and advice based on AI analysis of current conditions and user input.</li>


<h2>🚀 Quick Start</h2>

<h3>

[Live Demo]([https://safeguardian-llm.streamlit.app/](https://github.com/Ashoka74/SafeGuardian-LLM.git))
  
</h3>

  <li>Install dependencies:
    <pre><code>pip install -r requirements.txt</code></pre>
  </li>
  <li>Set up environment variables:
    Create a <code>.env</code> file in the project root with the following content:
    <pre><code>
     OPENAI_API = 'XXX'
     GROQ_API = 'XXX'
     HELICONE_API = 'XXX'
</code></pre>
  </li>
  
 
  <li>- Generate a .json access-key from FireBase</li>

   ![generate_key_firebase](https://github.com/user-attachments/assets/528aa756-6215-4ccc-b4e0-f716dd152f68)
  
  <li>- rename the file disasterrescueai-firebase-adminsdk.json</li>
  <li>Run the application:
    <pre><code>streamlit run app.py</code></pre>
  </li>
</ol>
