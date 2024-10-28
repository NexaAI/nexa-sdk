from typing import List, Any, Mapping, Optional
from langchain_core.embeddings import Embeddings
from langchain_core.pydantic_v1 import BaseModel, PrivateAttr
from nexa.gguf import NexaTextInference

class NexaEmbeddings(BaseModel, Embeddings):
    """Nexa embeddings using the NexaTextInference API.

    Example:
        .. code-block:: python

            from AMD_demo.nexa_embedding import NexaEmbeddings
            nexa_emb = NexaEmbeddings(
                model_path="nomic",
            )
            r1 = nexa_emb.embed_documents(
                [
                    "Alpha is the first letter of Greek alphabet",
                    "Beta is the second letter of Greek alphabet",
                ]
            )
            r2 = nexa_emb.embed_query(
                "What is the second letter of Greek alphabet"
            )
    """

    model_path: str
    """Path to the Nexa model."""

    embed_instruction: str = "passage: "
    """Instruction used to embed documents."""
    query_instruction: str = "query: "
    """Instruction used to embed the query."""

    _inference: NexaTextInference = PrivateAttr()

    def __init__(self, **data):
        super().__init__(**data)
        self._inference = NexaTextInference(model_path=self.model_path, embedding=True)

    @property
    def _identifying_params(self) -> Mapping[str, Any]:
        """Get the identifying parameters."""
        return {"model_path": self.model_path}

    def embed_documents(self, texts: List[str]) -> List[List[float]]:
        """Embed documents using the Nexa embedding model.

        Args:
            texts: The list of texts to embed.

        Returns:
            List of embeddings, one for each text.
        """
        embeddings = []
        for text in texts:
            embedding = self._inference.create_embedding(f"{self.embed_instruction}{text}")
            embeddings.append(embedding)
        return embeddings

    def embed_query(self, text: str) -> List[float]:
        """Embed a query using the Nexa embedding model.

        Args:
            text: The text to embed.

        Returns:
            Embeddings for the text.
        """
        return self._inference.create_embedding(f"{self.query_instruction}{text}")
