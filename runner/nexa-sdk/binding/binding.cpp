#include "binding.h"

#include <cstdio>
#include <cstring>
#include <string>
#include <vector>

#include "llama.h"

struct LLMPipeline {
  llama_model *model;
  const llama_vocab *vocab;
  llama_context *context;
  llama_sampler *sampler;
  // for chat
  std::vector<llama_chat_message> messages;
  std::vector<char> formatted;
  int prev_len = 0;
  // for streaming
  std::vector<llama_token> prompt_tokens;
  llama_token new_token_id;
  llama_batch batch;
  const char *tmpl;
  std::string response;
};

void init() { ggml_backend_load_all(); }

LLMPipeline *llm_pipeline_new() { return new LLMPipeline; }

void llm_pipeline_free(LLMPipeline *p) { delete p; }

bool llm_pipeline_load_model(LLMPipeline *p, char *model_path) {
  const int n_ctx = 40960;
  const float min_p = 0.00f;
  const float temp = 0.6f;

  // initialize the model
  llama_model_params model_params = llama_model_default_params();
  // model_params.n_gpu_layers = ngl;

  p->model = llama_model_load_from_file(model_path, model_params);
  if (!p->model) {
    fprintf(stderr, "%s: error: unable to load model\n", __func__);
    return false;
  }

  p->vocab = llama_model_get_vocab(p->model);

  // initialize the context
  llama_context_params ctx_params = llama_context_default_params();
  ctx_params.n_ctx = n_ctx;
  ctx_params.n_batch = n_ctx;

  p->context = llama_init_from_model(p->model, ctx_params);
  if (!p->context) {
    fprintf(stderr, "%s: error: failed to create the llama_context\n", __func__);
    return false;
  }

  // initialize the sampler
  p->sampler = llama_sampler_chain_init(llama_sampler_chain_default_params());
  llama_sampler_chain_add(p->sampler, llama_sampler_init_min_p(min_p, 1.0f));
  llama_sampler_chain_add(p->sampler, llama_sampler_init_temp(temp));
  llama_sampler_chain_add(p->sampler, llama_sampler_init_dist(LLAMA_DEFAULT_SEED));
  return true;
}

void llm_pipeline_close(LLMPipeline *p) {
  llama_sampler_free(p->sampler);
  llama_free(p->context);
  llama_model_free(p->model);
}

int32_t llm_pipeline_generate(LLMPipeline *p, char *user, char *res) {
  auto generate = [&](const std::string &prompt, std::string &response) {
    uint32_t token_count = 0;
    const bool is_first = llama_memory_seq_pos_max(llama_get_memory(p->context), 0) == 0;

    // tokenize the prompt
    const int n_prompt_tokens = -llama_tokenize(p->vocab, prompt.c_str(), prompt.size(), NULL, 0, is_first, true);
    std::vector<llama_token> prompt_tokens(n_prompt_tokens);
    if (llama_tokenize(p->vocab, prompt.c_str(), prompt.size(), prompt_tokens.data(), prompt_tokens.size(), is_first,
                       true) < 0) {
      GGML_ABORT("failed to tokenize the prompt\n");
    }

    // prepare a batch for the prompt
    llama_batch batch = llama_batch_get_one(prompt_tokens.data(), prompt_tokens.size());
    llama_token new_token_id;
    while (true) {
      // check if we have enough space in the context to evaluate this batch
      int n_ctx = llama_n_ctx(p->context);
      int n_ctx_used = llama_memory_seq_pos_max(llama_get_memory(p->context), 0);
      if (n_ctx_used + batch.n_tokens > n_ctx) {
        fprintf(stderr, "context size exceeded\n");
        exit(0);
      }

      if (llama_decode(p->context, batch)) {
        GGML_ABORT("failed to decode\n");
      }

      // sample the next token
      new_token_id = llama_sampler_sample(p->sampler, p->context, -1);

      // is it an end of generation?
      if (llama_vocab_is_eog(p->vocab, new_token_id)) {
        break;
      }

      // convert the token to a string, print it and add it to the response
      char buf[256];
      int n = llama_token_to_piece(p->vocab, new_token_id, buf, sizeof(buf), 0, true);
      if (n < 0) {
        GGML_ABORT("failed to convert token to piece\n");
      }
      std::string piece(buf, n);
      printf("%s", piece.c_str());
      fflush(stdout);
      response += piece;
      token_count += 1;

      // prepare the next batch with the sampled token
      batch = llama_batch_get_one(&new_token_id, 1);
    }

    return token_count;
  };

  const char *tmpl = llama_model_chat_template(p->model, /* name */ nullptr);

  // add the user input to the message list and format it
  p->messages.push_back({"user", strdup(user)});
  int new_len = llama_chat_apply_template(tmpl, p->messages.data(), p->messages.size(), true, p->formatted.data(),
                                          p->formatted.size());
  if (new_len > (int)p->formatted.size()) {
    p->formatted.resize(new_len);
    new_len = llama_chat_apply_template(tmpl, p->messages.data(), p->messages.size(), true, p->formatted.data(),
                                        p->formatted.size());
  }
  if (new_len < 0) {
    fprintf(stderr, "failed to apply the chat template\n");
    return -1;
  }

  // remove previous messages to obtain the prompt to generate the response
  std::string prompt(p->formatted.begin() + p->prev_len, p->formatted.begin() + new_len);

  // generate a response
  std::string response;
  uint32_t token_count = generate(prompt, response);
  if (response.size() > 65535) {
    fprintf(stderr, "response too large\n");
    return -1;
  }
  std::strncpy(res, response.data(), response.size());

  // add the response to the messages
  p->messages.push_back({"assistant", strdup(response.c_str())});
  p->prev_len = llama_chat_apply_template(tmpl, p->messages.data(), p->messages.size(), false, nullptr, 0);
  if (p->prev_len < 0) {
    fprintf(stderr, "failed to apply the chat template\n");
    return -1;
  }

  return token_count;
}

bool llm_pipeline_generate_send(LLMPipeline *p, char *user) {
  p->tmpl = llama_model_chat_template(p->model, /* name */ nullptr);
  p->messages.push_back({"user", strdup(user)});
  int new_len = llama_chat_apply_template(p->tmpl, p->messages.data(), p->messages.size(), true, p->formatted.data(),
                                          p->formatted.size());
  if (new_len > (int)p->formatted.size()) {
    p->formatted.resize(new_len);
    new_len = llama_chat_apply_template(p->tmpl, p->messages.data(), p->messages.size(), true, p->formatted.data(),
                                        p->formatted.size());
  }
  if (new_len < 0) {
    fprintf(stderr, "failed to apply the chat template\n");
    return false;
  }

  // remove previous messages to obtain the prompt to generate the response
  std::string prompt(p->formatted.begin() + p->prev_len, p->formatted.begin() + new_len);

  const bool is_first = llama_memory_seq_pos_max(llama_get_memory(p->context), 0) == 0;

  // tokenize the prompt
  const int n_prompt_tokens = -llama_tokenize(p->vocab, prompt.c_str(), prompt.size(), NULL, 0, is_first, true);
  std::vector<llama_token> prompt_tokens(n_prompt_tokens);
  p->prompt_tokens = prompt_tokens;
  if (llama_tokenize(p->vocab, prompt.c_str(), prompt.size(), p->prompt_tokens.data(), p->prompt_tokens.size(),
                     is_first, true) < 0) {
    GGML_ABORT("failed to tokenize the prompt\n");
  }

  // prepare a batch for the prompt
  p->response.clear();
  p->batch = llama_batch_get_one(p->prompt_tokens.data(), p->prompt_tokens.size());

  return true;
}

int32_t llm_pipeline_generate_next_token(struct LLMPipeline *p, char *res) {
  // check if we have enough space in the context to evaluate this batch
  int n_ctx = llama_n_ctx(p->context);
  int n_ctx_used = llama_memory_seq_pos_max(llama_get_memory(p->context), 0);
  if (n_ctx_used + p->batch.n_tokens > n_ctx) {
    fprintf(stderr, "context size exceeded\n");
    exit(0);
  }

  if (llama_decode(p->context, p->batch)) {
    GGML_ABORT("failed to decode\n");
  }

  // sample the next token
  p->new_token_id = llama_sampler_sample(p->sampler, p->context, -1);

  // is it an end of generation?
  if (llama_vocab_is_eog(p->vocab, p->new_token_id)) {
    p->messages.push_back({"assistant", strdup(p->response.c_str())});
    p->prev_len = llama_chat_apply_template(p->tmpl, p->messages.data(), p->messages.size(), false, nullptr, 0);
    if (p->prev_len < 0) {
      fprintf(stderr, "failed to apply the chat template\n");
      return -1;
    }
    return 0;
  }

  // convert the token to a string, print it and add it to the response
  int n = llama_token_to_piece(p->vocab, p->new_token_id, res, 256, 0, true);
  if (n < 0) {
    GGML_ABORT("failed to convert token to piece\n");
  }
  res[n] = '\0';
  p->response.append(res, n);

  // prepare the next batch with the sampled token
  p->batch = llama_batch_get_one(&p->new_token_id, 1);

  return n;
}
