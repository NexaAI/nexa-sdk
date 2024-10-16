from multiprocessing.connection import Connection


def sync_with_parent(child_connection: Connection) -> None:
    child_connection.recv()
    child_connection.send(0)


def sync_with_child(parent_connection: Connection) -> None:
    parent_connection.send(0)
    parent_connection.recv()
